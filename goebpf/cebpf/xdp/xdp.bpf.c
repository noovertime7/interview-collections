//go:build ignore
#include <common.h>
#include <linux/tcp.h>
#include <bpf_endian.h>

// 定义一个结构体, 用于存储IP地址
struct data_ip {
    __u32 sip; //源IP地址
    __u32 dip; //目的IP地址
    __u32 pkt_sz; //数据包大小
    __u32 iii; //ingress_ifindex 网卡索引
    __be16 sport; //源端口
    __be16 dport; //目的端口
};

//ringbuf
struct { //ringbuf，环形缓冲区，算是一种用户内核交互的优先选择
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries,1<<20); //大概是10M大小
} ip_map SEC(".maps");

 // 存放ip白名单的hashmap，数据由用户态塞入
struct bpf_map_def SEC("maps") allow_ips_map = {
     .type = BPF_MAP_TYPE_HASH,
     .key_size = sizeof(__u32), //匹配ip地址类型
     .value_size = sizeof(__u8), //值为1字节，用于标记是否允许
     .max_entries = 1024,
 };

SEC("xdp")
int my_pass(struct xdp_md *ctx)
{
    void* data = (void*)(long)ctx->data;
    void* data_end = (void*)(long)ctx->data_end;
    int pkt_sz = data_end - data;

    //由于数据报文是通过不同层的协议封装的，所以我们需要逐层解析，取出每层的头部信息
    //最外层为数据链路层，所以data指向的是ethernet头部

    struct ethhdr *eth = data; // 链路层
    if ((void*)eth + sizeof(*eth) > data_end) { //如果包不完整，或者数据被篡改，直接丢弃
        bpf_printk("Invalid ethernet header\n");
        return XDP_DROP;
    }

    struct iphdr *ip = data + sizeof(*eth); // 网络层
    if ((void*)ip + sizeof(*ip) > data_end) {
        bpf_printk("Invalid ip header\n");
        return XDP_DROP;
    }
    if (ip->protocol != 6) { //如果不是TCP 就不处理了
        return XDP_PASS;
    }

    struct tcphdr *tcp = (void*)ip + sizeof(*ip); // 传输层
    if ((void*)tcp + sizeof(*tcp) > data_end) {
        bpf_printk("Invalid tcp header\n");
        return XDP_DROP;
    }

    //从ip层数据获取源ip地址
//    __u32 src_ip = ip->saddr;
//    bpf_printk("Source IP address is %u.%u.%u.",
//               src_ip & 0xFF, (src_ip >> 8) & 0xFF,
//               (src_ip >> 16) & 0xFF); //临时代码，类似的业务处理是发送到用户态完成的
//    //ebpf程序存在一定的限制，不能直接打印四个字节的ip地址，也不能使用标准库
//    bpf_printk("%u\n", (src_ip >> 24) & 0xFF);

    //这里打印protocol是数字，每个数字对应一个协议，如1对应ICMP，6对应TCP，17对应UDP
//    bpf_printk("output:Packet size is %d, protocol is %d\n",
//     ctx->data_end - ctx->data,
//     ip->protocol);

    struct data_ip *ipdata;
    ipdata=bpf_ringbuf_reserve(&ip_map, sizeof(*ipdata), 0); //在ringbuf中预留缓冲区大小
    if(!ipdata){
      return 0;
    }
    ipdata->sip=bpf_ntohl(ip->saddr); //将网络字节序转换为主机字节序 32位
    ipdata->dip=bpf_ntohl(ip->daddr);
    ipdata->pkt_sz=pkt_sz;
    ipdata->iii=ctx->ingress_ifindex;
    ipdata->sport=bpf_ntohl(tcp->source); //16位
    ipdata->dport=bpf_ntohl(tcp->dest);

    bpf_ringbuf_submit(ipdata, 0); //将ringbuf数据发送到用户态

    //获取从用户态发送回的ip白名单
    __u32 key = bpf_ntohl(ip->saddr);
    __u8 *allow = bpf_map_lookup_elem(&allow_ips_map, &key);
    if (allow && *allow == 1){
        return XDP_PASS; //放行
    }

    return XDP_DROP; //什么都不做 就是放行
}

char _license[] SEC("license") = "GPL";
