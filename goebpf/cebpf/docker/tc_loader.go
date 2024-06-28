package docker

import (
	"errors"
	"fmt"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/vishvananda/netlink"
	"goebpf/pkg/helpers/nethelper"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

type TcDataIp struct { //对应mydockertc.bpf.c中的struct
	Sip   uint32
	Dip   uint32
	Sport uint32
	Dport uint32
}

func ClearTC() {
	veth := nethelper.GetVeths()
	for _, v := range veth {
		err := netlink.QdiscDel(&netlink.GenericQdisc{
			QdiscAttrs: netlink.QdiscAttrs{
				LinkIndex: v.Index,
			},
		})
		if err != nil {
			fmt.Println("QdiscDel err: ", err.Error())
		}
	}
}

// 在目标网卡添加clsact队列，使其成为eBPF监听的对象,来源——cillium源码
// https://github.com/cilium/cilium/blob/main/pkg/datapath/loader/tc.go
func attachIface(linkIndex int, fd int, name string) (deferFuncs []func()) {
	//2.1初始化队列
	attrs := netlink.QdiscAttrs{
		LinkIndex: linkIndex,
		// 0xffff 表示 “根”或“无父”句柄的队列规则
		Handle: netlink.MakeHandle(0xffff, 0),
		Parent: netlink.HANDLE_CLSACT, //eBPF专用 clsact
	}
	qdisc := &netlink.GenericQdisc{
		QdiscAttrs: attrs,
		QdiscType:  "clsact",
	}
	//2.2添加队列 —— 好比执行了 tc qdisc add dev docker0  clsact
	if err := netlink.QdiscAdd(qdisc); err != nil {
		log.Fatalln("QdiscAdd err: ", err)
	}
	deferFuncs = append(deferFuncs, func() { //监测完删除，否则下次无法创建
		if err := netlink.QdiscDel(qdisc); err != nil {
			fmt.Println("QdiscDel err: ", err.Error())
		}
	})

	//3.1初始化 eBPF分类器
	filterattrs := netlink.FilterAttrs{
		LinkIndex: linkIndex,
		Parent:    netlink.HANDLE_MIN_INGRESS | netlink.HANDLE_MIN_EGRESS,
		Handle:    netlink.MakeHandle(0, 1),
		Protocol:  unix.ETH_P_ALL, //所有协议
		Priority:  1,
	}
	filter := &netlink.BpfFilter{
		FilterAttrs:  filterattrs,
		Fd:           fd,
		Name:         name,
		DirectAction: true,
	}
	//3.2添加分类器 —— 好比执行了 tc filter add dev docker0 ingress bpf direct-action obj dockertcxdp_bpfel_x86.o
	if err := netlink.FilterAdd(filter); err != nil {
		log.Fatalln("FilterAdd err: ", err)
	}
	deferFuncs = append(deferFuncs, func() {
		err := netlink.FilterDel(filter)
		if err != nil {
			fmt.Println("FilterDel err : ", err.Error())
		}
	})
	return
}

// 加载tc ebpf 程序
func LoaderTC() {
	veth := nethelper.GetVeths()

	//1 这步和其他的eBPF程序一样，加载转化过来的eBPF程序
	objs := &mydockertcObjects{}
	err := loadMydockertcObjects(objs, nil)
	if err != nil {
		log.Fatalln("loadDockertcxdpObjects err: ", err)
	}

	//2-3 给所有veth网卡添加clsact队列
	for _, v := range veth {
		deferFuncs := attachIface(v.Index, objs.Mytc.FD(), "mytc")
		for _, f := range deferFuncs {
			defer f()
		}
	}

	//4开个信号阻塞住并循环读取
	fmt.Println("开始TC监听")
	go func() {
		rd, err := ringbuf.NewReader(objs.TcIpMap)
		if err != nil {
			log.Fatalf("creating event reader: %s", err)
		}
		defer rd.Close()
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
				data := (*TcDataIp)(unsafe.Pointer(&record.RawSample[0])) //经过两次强制转换

				//转换成网络字节序
				saddr := nethelper.ResolveIP(data.Sip, true)
				daddr := nethelper.ResolveIP(data.Dip, true)

				fmt.Printf("监测到来源地址: %s:%d------->目标地址: %s:%d\n",
					saddr.To4().String(), data.Sport,
					daddr.To4().String(), data.Dport,
				)
			}
		}
	}() //循环读取内核态传来的数据
	//开信号 好处是能执行defer
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGHUP)
	<-ch
	fmt.Println("TC监听结束")
}
