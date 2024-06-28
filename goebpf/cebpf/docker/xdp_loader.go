package docker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"goebpf/pkg/helpers/dbhelper"
	"goebpf/pkg/helpers/httphelper"
	"goebpf/pkg/helpers/nethelper"
	"log"
	"net"
	"unsafe"
)

//用于处理容器网络监测，包含xdp，tc

type DataIp struct { //定义一个与bpf中的struct对应的go struct
	Sip     uint32
	Dip     uint32
	Sport   uint16
	Dport   uint16
	Payload [1024]byte //报文
}

// 向map中添加允许的IP，这里正确的姿势是从缓存中读取ip list，而非读取文件或从数据库中读取
func allowIP(set *ebpf.Map) {
	//添加允许的IP
	ip1 := binary.BigEndian.Uint32(net.ParseIP("172.17.0.2").To4())
	if err := set.Put(ip1, uint8(1)); err != nil {
		log.Fatalf("can't update allow ips: %v", err)
	}
}

// 中间函数，加载bpf程序供main调用
func LoaderVethXDP() {
	docker_obj := mydockerObjects{}
	err := loadMydockerObjects(&docker_obj, nil) //加载bpf对象
	if err != nil {
		log.Fatalf("can't load myxdp: %v", err)
	}
	defer docker_obj.Close()

	//加载ip白名单
	//allowIP(xdp_obj.AllowIpsMap)

	//获取所有veth网卡
	veths := nethelper.GetVeths()
	for _, veth := range veths { //把所有veth都绑定xdp
		xdp, err := link.AttachXDP(link.XDPOptions{
			Interface: veth.Index,
			Program:   docker_obj.DockerNet,
		})
		if err != nil {
			log.Fatalf("can't attach xdp: %v", err)
		}
		defer xdp.Close()
	}

	//创建reader用来读取内核map
	rd, err := ringbuf.NewReader(docker_obj.IpMap)
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
			saddr := nethelper.ResolveIP(data.Sip, true)
			daddr := nethelper.ResolveIP(data.Dip, true)

			comment := "none"
			if req, ok := httphelper.IsHttpRequest(data.Payload[:]); ok {
				comment = fmt.Sprintf("HTTP 请求: %s %s %s",
					req.Method, req.URL.Path, req.URL.RawQuery)
			} else if resp, ok := httphelper.IsHttpResponse(data.Payload[:]); ok {
				comment = fmt.Sprintf("HTTP 响应: %s", resp.Status)
			} else if sql, err := dbhelper.ExtractSQLFromMySQLPacket(data.Payload[:]); err == nil {
				comment = fmt.Sprintf("MySQL 查询: %s", sql)
			} else if rds, err := dbhelper.ExtractSQLFromRedisPacket(data.Payload[:]); err == nil {
				comment = fmt.Sprintf("Redis: %s", rds)
			}

			fmt.Printf("监测到来源IP: %s------->目标IP: %s,说明:%s\n",
				saddr.To4().String(),
				daddr.To4().String(),
				comment)
		}
	}
}
