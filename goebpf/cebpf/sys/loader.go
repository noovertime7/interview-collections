package sys

import (
	"errors"
	"fmt"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"golang.org/x/sys/unix"
	"log"
	"time"
	"unsafe"
)

//啥都有，tracepoint，kprobe，uprobe

type ProcT struct { //定义一个与bpf中的struct对应的go struct
	Pid  uint32
	Ppid uint32
	Comm [256]byte
}

// LoaderProc tracepoint 加载了sys_exit_execve事件，execve是用新建进程替换当前进程信息的事件，譬如终端执行ls，实际是替换了原bash进程
func LoaderProc() {
	sys_obj := &sysObjects{}
	err := loadSysObjects(sys_obj, nil)
	if err != nil {
		log.Fatalf("can't load sys: %v", err)
	}
	defer sys_obj.Close()

	tp, err := link.Tracepoint("syscalls", "sys_exit_execve", sys_obj.HandleProc, nil)
	if err != nil {
		log.Fatalf("can't attach tracepoint: %v", err)
	}
	defer tp.Close()

	rd, err := ringbuf.NewReader(sys_obj.ProcMap)
	if err != nil {
		log.Fatalf("creating event reader: %s", err)
	}
	defer rd.Close()

	fmt.Println("进程监控开始")
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

		if len(record.RawSample) > 0 {
			proc := (*ProcT)(unsafe.Pointer(&record.RawSample[0])) //经过两次强制转换
			fmt.Printf("监测到进程启动，进程名: %s，进程id: %d，进程父id: %d\n",
				string(proc.Comm[:]), //需要转换成切片
				proc.Pid, proc.Ppid)
		}
	}
}

// LoaderTaskSwitch kprobe 用于监测进程切换，指cpu上下文切换
func LoaderTaskSwitch() {
	sys_obj := &sysObjects{}
	err := loadSysObjects(sys_obj, nil)
	if err != nil {
		log.Fatalf("can't load sys: %v", err)
	}
	defer sys_obj.Close()
	//这里是通过命令查找的内核函数，用法在readme中有说明
	tp, err := link.Kprobe("finish_task_switch.isra.0", sys_obj.HandleTaskSwitch, nil)
	if err != nil {
		log.Fatalf("can't attach tracepoint: %v", err)
	}
	defer tp.Close()

	for { //临时用的
		time.Sleep(1 * time.Second)
	}
}

// 监测用户函数调用
type Event struct {
	Pid  uint32
	Line [80]uint8
}

const (
	//这套组合用于获取当前输入命令 ，用于审计
	binPath        = "/bin/bash"
	readlineSymbol = "readline" //nm -D /bin/bash |grep readline -D 代表查询动态链接库的一些函数
	//这套组合就是自编go程序，获取其中一个函数的返回值， 用于业务日志
	myappPath   = "/home/ubuntu/app/goebpf/test/testapp"
	myappSymbol = "main.myHandler" //nm testapp |grep myHandler nm表示用来查询当前可执行程序的符号表
)

// LoaderUretprobe uprobe
func LoaderUretprobe() { //从 https://github.com/cilium/ebpf/blob/main/examples/uretprobe/ 抄的
	//初始化ebpf对象
	sys_obj := &sysObjects{}
	err := loadSysObjects(sys_obj, nil)
	if err != nil {
		log.Fatalf("can't load sys: %v", err)
	}
	defer sys_obj.Close()
	//打开可执行程序
	ex, err := link.OpenExecutable(myappPath)
	if err != nil {
		log.Fatalf("opening executable: %s", err)
	}
	//可执行程序里面的用户函数
	up, err := ex.Uretprobe(myappSymbol, sys_obj.UretprobeBashReadline, nil)
	if err != nil {
		log.Fatalf("creating uretprobe: %s", err)
	}
	defer up.Close()
	//打开reader
	rd, err := ringbuf.NewReader(sys_obj.EventMap)
	if err != nil {
		log.Fatalf("creating event reader: %s", err)
	}
	defer rd.Close()

	fmt.Println("用户函数调用监控开始")
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

		if len(record.RawSample) > 0 {
			event := (*Event)(unsafe.Pointer(&record.RawSample[0])) //经过两次强制转换
			fmt.Printf("监测到用户函数调用，进程id: %d，返回结果: %s\n",
				event.Pid, unix.ByteSliceToString(event.Line[:]))
		}
	}
}
