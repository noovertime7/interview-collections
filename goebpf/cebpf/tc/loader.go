package tc

import (
	"errors"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"log"
	"unsafe"
)

type DataT struct { //定义一个与bpf中的struct对应的go struct
	Pid  uint32
	Comm [255]byte //数组
}

// 中间函数，执行bpf生成.go文件里的代码供main函数调用
// 这里一系列内存以及tracepoint操作需要在root权限下进行
func LoaderTcWrite() {
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	} //移除内存锁定限制 细节操作 不重要

	tc_obj := tc_writeObjects{}
	err := loadTc_writeObjects(&tc_obj, nil) //加载bpf对象
	if err != nil {
		log.Fatalf("can't load tc_write: %v", err)

	}

	//抄的，将tracepoint指向tc_obj.HandleTp
	tp, err := link.Tracepoint("syscalls", "sys_enter_openat", tc_obj.HandleTp, nil)
	if err != nil {
		log.Fatalf("can't attach tracepoint: %v", err)
	}
	defer tp.Close()

	//创建reader用来读取内核map
	//rd, err := perf.NewReader(tc_obj.MyBpfMap, os.Getpagesize())
	rd, err := ringbuf.NewReader(tc_obj.LogMap)
	if err != nil {
		log.Fatalf("creating event reader: %s", err)
	}
	defer rd.Close()

	for { //循环读取内核map
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, perf.ErrClosed) {
				log.Println("Received signal, exiting..")
				return
			}
			log.Printf("reading from reader: %s", err)
			continue
		}

		if len(record.RawSample) > 0 {
			data := (*DataT)(unsafe.Pointer(&record.RawSample[0]))   //经过两次强制转换
			log.Printf("监测到进程正在写文件，进程名: %s\n", string(data.Comm[:])) //需要转换成切片
		}
	}

}
