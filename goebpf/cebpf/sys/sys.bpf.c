//go:build ignore
//通过bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h 生成
#include <vmlinux.h> // 包含了系统运行Linux 内核源代码中使使用的所有类型定义
#include <bpf_helpers.h>
#include <bpf_tracing.h>
//#include <common.h>

// 可以理解为操作内核的一个凭证
char LICENSE[] SEC("license") = "Dual BSD/GPL";

struct proc_t{
    __u32 pid;
    __u32 ppid;//父进程id
    char pname[256];
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries,1<<20);
} proc_map SEC(".maps");

//创建新进程结束的事件
SEC("tracepoint/syscalls/sys_exit_execve")
int handle_proc(void *ctx) {
    //在ringbuf中预留数据空间
    struct proc_t *p = NULL;
    p=bpf_ringbuf_reserve(&proc_map, sizeof(*p), 0);
    if(!p){
      return 0;
    }

    //获取进程信息
    p->pid = bpf_get_current_pid_tgid() >> 32;
    bpf_get_current_comm(&p->pname, sizeof(p->pname));

    //获取父进程信息，需要读取内核数据
    p->ppid=0;
    struct task_struct *task=(struct task_struct *)bpf_get_current_task(); //内核源代码中的类型
    if(task){ // 获取进程信息 OK
        struct task_struct *parent=NULL;
        //因为安全限制，保护内核地址空间，无法直接读取task的数据
        //需要使用bpf_helper中的方法来读取
        bpf_probe_read_kernel(&parent,sizeof(parent),&task->real_parent);
        if(parent){//当前进程有父进程
            bpf_probe_read_kernel(&p->ppid,sizeof(p->ppid),&parent->pid);
        }
    }
    //提交数据
    bpf_ringbuf_submit(p, 0);
    return 0;
}

//进程切换
SEC("kprobe/finish_task_switch")
int handle_task_switch(struct task_struct *prev){
    u32 cur_pid = 0;
    u32 prev_pid = 0;
    struct task_struct *cur = (struct task_struct *)bpf_get_current_task();
    if (cur) {
        bpf_probe_read_kernel(&cur_pid, sizeof(cur_pid), &cur->pid);
    }
    if (prev) {
        bpf_probe_read_kernel(&prev_pid, sizeof(prev_pid), &prev->pid);
    }
    if (prev_pid !=0){
        bpf_printk("prev_pid:%d, cur_pid:%d\n", prev_pid, cur_pid);
    }
    return 0;
}

//用户函数调用
//从https://github.com/cilium/ebpf/blob/main/examples/uretprobe/ 抄的
struct event {
	u32 pid;
	u8 line[80];
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries,1<<20);
} event_map SEC(".maps");

// = user return probe 代表获取用户函数调用的返回值
SEC("uretprobe/abc") //这里的名字是随意的
int uretprobe_bash_readline(struct pt_regs *ctx) {
    //在ringbuf中预留数据空间
    struct event *ev = NULL;
    ev=bpf_ringbuf_reserve(&event_map, sizeof(*ev), 0);
    if(!ev){
      return 0;
    }

    //获取数据
    //PT_REGS_RC用于获取返回值，需要在编译时指定cpu架构
    ev->pid = bpf_get_current_pid_tgid();
    bpf_probe_read(&ev->line, sizeof(ev->line), (void *)PT_REGS_RC(ctx));

    //提交数据
    bpf_ringbuf_submit(ev, 0);

    return 0;
}


