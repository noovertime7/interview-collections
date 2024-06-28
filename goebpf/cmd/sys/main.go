package main

import "goebpf/cebpf/sys"

func main() {
	//sys.LoaderProc() //监控进程启动
	//sys.LoaderTaskSwitch() //监控进程切换
	sys.LoaderUretprobe() //监控用户函数调用
}
