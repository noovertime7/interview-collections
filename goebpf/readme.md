## 使用cilium对ebpf的用户态进行调用

# 官方的示例代码 https://github.com/cilium/ebpf/blob/master/examples/tracepoint_in_go/main.go

# 安装工具
go get  github.com/cilium/ebpf/cmd/bpf2go
go install  github.com/cilium/ebpf/cmd/bpf2go
需要添加/go/bin到环境变量中，用于执行生成的工具文件
这是转换程序，允许在Go 代码中编译和嵌入eBPF 程序

# 安装依赖包
sudo apt install llvm
sudo apt install clang

# 安装C依赖库
sudo apt install libelf-dev
git clone --depth 1 https://github.com/libbpf/libbpf
cd src
make install
~~sudo ln -s /usr/include/asm-generic  /usr/include/asm~~ （这个库运行报错，拿软链接尝试解决了问题）
sudo ln -s /usr/include/x86_64-linux-gnu/asm /usr/include/asm

# 使用方法
1.项目目录下执行make把操作内核态的c文件编译生成.go和.o文件                 cmd:make->Makefile->doc.go
2.编写方法来加载bpf program对象，并将tracepoint指向它，供main函数调用      loader.go->tc_write_bpfel.go
3.go run cmd/tc/main.go运行                                           main.go->loader.go
4.执行完毕后，可以通过cat /sys/kernel/debug/tracing/trace_pipe查看输出(仅限bpf_printk)

tips:
1.通过perf list|grep sys_exit_execve 查看具体的tracepoint
2.通过cat /sys/kernel/debug/tracing/available_filter_functions|grep finish_task_switch 查看具体的kprobe（这里的名称用于用户态去link）
3.如需读取内核数据，如获取父进程pid，可以执行
bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h
包含了系统运行Linux 内核源代码中使使用的所有类型定义
4.https://github.com/torvalds/linux 查看源码获取内核函数的签名


# 目录结构
.
├── Makefile `用来加载环境变量并执行编译`
├── cebpf
│        ├── headers `用于存放bpf相关的头文件` downloaded from https://github.com/cilium/ebpf/tree/main/examples/headers
│        └── tc `"tracepoint`
│            ├── doc.go `实际的编译命令存放的地方，通过makefile来指向`
│            ├── loader.go `用于将tracepoint指向加载了的bpf程序`
│            ├── tc_write.bpf.c         `原始bpf代码`
│            ├── tc_write_bpfeb.go      `---`
│            ├── tc_write_bpfeb.o       `⬆️ 这些全是编译生成的文件`
│            ├── tc_write_bpfel.go      `⬇️ 包含了所要加载的bpf程序对象`
│            └── tc_write_bpfel.o       `----`
│        └── xdp `xdp相关的用户态操作，都是网络相关，相比tracepoint只有少数的调用需要修改`
│        └── sys `用于进行进程相关的用户态操作，写的比较杂，tp，kprobe，uprobe都有`
│        └── docker `容器间网络互访，包含了xdp，tc`
├── cmd
│        └── tc
│            └── main.go `主函数入口`

### 一些概念分类上的问题
tc(traffic control)与xdp都是网络相关的 
区别在于：xdp是作用与设备驱动上，而tc是作用在linux流量控制器上
tc是本身存在的，因此只需要创建一个clsact类型的队列作为程序挂载的入口，就像hook一样
tc可以更方便地修改报文，端口，地址等

### traffic control入门——命令行方式加载bpf程序
1.编译
2.tc qdisc add dev docker0 clsact ---使用docker0创建一个队列
3.tc filter add dev docker0 ingress bpf direct-action obj mydockertc_x86_bpfel.o 

清理掉
tc qdisc del dev docker0 clsact
tc qdisc del dev vethd5577c5@if30 clsact
tc qdisc del dev vethe461882@if46 clsact
查看
tc filter show dev vethd5577c5 ingress



