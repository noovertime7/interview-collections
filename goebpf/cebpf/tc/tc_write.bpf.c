//go:build ignore
#define BPF_NO_GLOBAL_DATA
//#include <linux/bpf.h>
//#include <bpf/bpf_helpers.h>
//#include <bpf/bpf_tracing.h>
#include <common.h> //替代了库里面的bpf头文件
#include <linux/limits.h>
typedef unsigned int u32;

char LICENSE[] SEC("license") = "Dual BSD/GPL";

int is_eq(char *str1,char *str2){
    int eq=1;
     int i ;
     for (i=0;i<sizeof(str1)-1 && i<sizeof(str2)-1;i++){
         if (str1[i]!=str2[i]){
            eq=0;
            break;;
         }
     }
     return eq;
}
// 定义一个结构体，用于pid和进程名称
struct data_t {
    u32 pid;
    char comm[NAME_MAX];  //NAME_MAX 文件名的最大长度，通常也可以用于进程或线程名称的最大长度
};

//ebpf提供了多种用户和内核交互的方式，用以满足不同场景的需求
//譬如ringbuf就能解决内存效率和事件重排序问题

//1.perf
//struct bpf_map_def SEC("maps") my_bpf_map = { //通过语法糖创建一个用于和用户态交互数据的map
//  .type       = BPF_MAP_TYPE_PERF_EVENT_ARRAY, //各种类型 看业务场景
//  .key_size   = sizeof(int),
//  .value_size   = sizeof(int),
//  .max_entries = 0, //这里指的是用户态发送过来的数据最大限制
//};

//2.ringbuf
struct { //ringbuf，环形缓冲区，算是一种用户内核交互的优先选择
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries,1<<20); //大概是10M大小
} log_map SEC(".maps");

SEC("tracepoint/syscalls/sys_enter_write")
int handle_tp(void *ctx)
{
//    char app_name[]="testwrite";  //这是一个全局变量，编译时会有警告

    //perf获取pid和进程名称的方式
//    struct data_t data = {};
//    data.pid = bpf_get_current_pid_tgid() >> 32; //获取PID
//    bpf_get_current_comm(&data.comm, sizeof(data.comm)); //获取进程名称

    //ringbuf获取内核数据要稍作改变
    struct data_t *data;
    data=bpf_ringbuf_reserve(&log_map, sizeof(*data), 0); //在ringbuf中预留缓冲区大小
    if(!data){
      return 0;
    }
    data->pid = bpf_get_current_pid_tgid() >> 32; //获取PID
    bpf_get_current_comm(&data->comm, sizeof(data->comm)); //获取进程名称

//     int eq=is_eq(data.comm,app_name);
//    if(eq==1){
//       bpf_printk("pid= %d,name:%s. writing data\n",  data.pid, data.comm); //临时测试，会打印到/sys/kernel/debug/tracing/trace_pipe
//       bpf_perf_event_output(ctx, &my_bpf_map, BPF_F_CURRENT_CPU, &data, sizeof(data)); //将perf_event发送到用户态
         bpf_ringbuf_submit(data, 0); //将ringbuf数据发送到用户态
//    }
   return 0;
}
