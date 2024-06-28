package xdp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"log"
	"net"
	"unsafe"
)

//xdp，网络相关

type DataIp struct { //定义一个与bpf中的struct对应的go struct
	Sip   uint32
	Dip   uint32
	PktSz uint32
	III   uint32
	Sport uint16
	Dport uint16
}

func ntohs(port uint16) uint16 {
	return ((port & 0xff) << 8) | (port >> 8)
}

func resolveIP(input_ip uint32, isbig bool) net.IP {
	ipNetworkOrder := make([]byte, 4)
	if isbig {
		binary.BigEndian.PutUint32(ipNetworkOrder, input_ip)
	} else {
		binary.LittleEndian.PutUint32(ipNetworkOrder, input_ip)
	}

	return ipNetworkOrder
}

// 向map中添加允许的IP，这里正确的姿势是从缓存中读取ip list，而非读取文件或从数据库中读取
func allowIP(set *ebpf.Map) {
	//添加允许的IP
	ip1 := binary.BigEndian.Uint32(net.ParseIP("172.17.0.2").To4())
	if err := set.Put(ip1, uint8(1)); err != nil {
		log.Fatalf("can't update allow ips: %v", err)
	}
}

// 中间函数，执行bpf生成.go文件里的代码供main函数调用
func LoaderTcWrite() {
	xdp_obj := myxdpObjects{}
	err := loadMyxdpObjects(&xdp_obj, nil) //加载bpf对象
	if err != nil {
		log.Fatalf("can't load myxdp: %v", err)
	}
	defer xdp_obj.Close()

	//加载ip白名单
	allowIP(xdp_obj.AllowIpsMap)

	//获取网卡
	iface, err := net.InterfaceByName("docker0")
	//这里宿主机向容器访问，由于存在nat转换，所以来源ip会变成容器的ip
	if err != nil {
		log.Fatalf("can't get interface: %v", err)
	}

	//绑定xdp到指定网卡
	xdp, err := link.AttachXDP(link.XDPOptions{
		Interface: iface.Index,
		Program:   xdp_obj.MyPass,
	})
	if err != nil {
		log.Fatalf("can't attach xdp: %v", err)
	}
	defer xdp.Close()

	//创建reader用来读取内核map
	rd, err := ringbuf.NewReader(xdp_obj.IpMap)
	if err != nil {
		log.Fatalf("creating event reader: %s", err)
	}
	defer rd.Close()

	fmt.Println("开始监听IP包")
	for { //循环读取内核map
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				log.Println("Received signal, exiting..")
				return
			}
			log.Printf("reading from reader: %s", err)
			continue
		}

		//对内核态传来的数据进行解析
		if len(record.RawSample) > 0 {
			data := (*DataIp)(unsafe.Pointer(&record.RawSample[0])) //经过两次强制转换

			//转换成网络字节序
			saddr := resolveIP(data.Sip, true)
			daddr := resolveIP(data.Dip, true)

			fmt.Printf("监测到入口网卡index:%d,来源IP端口: %s:%d,目标IP端口: %s:%d,包大小: %d\n",
				data.III,
				saddr.To4().String(), data.Sport,
				daddr.To4().String(), data.Dport,
				data.PktSz)
			//指的一提的是
		}
	}
}
