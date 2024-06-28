//go:build ignore
#include <common.h>
#include <linux/tcp.h>
#include <bpf_endian.h>

#define HTTP_PAYLOAD_MAX 1024
// 定义一个结构体, 用于存储IP地址
struct data_ip {
    __u32 sip; //源IP地址
    __u32 dip; //目的IP地址
    __be16 sport; //源端口
    __be16 dport; //目的端口
    char payload[HTTP_PAYLOAD_MAX]; //HTTP数据
};

//ringbuf
struct { //ringbuf，环形缓冲区，算是一种用户内核交互的优先选择
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries,1<<20); //大概是10M大小
} ip_map SEC(".maps");

SEC("xdp")
int docker_net(struct xdp_md *ctx)
{
    void* data = (void*)(long)ctx->data;
    void* data_end = (void*)(long)ctx->data_end;

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

    //计算TCP数据长度，公式。抄的
    __be16 tcp_data_len = bpf_ntohs(ip->tot_len) - (ip->ihl*4) - (tcp->doff*4); //ip的报文=tcp报头+报文
    if (tcp_data_len == 0){ //数据包为0，则判断为tcp握手挥手，放行
        return XDP_PASS;
    }

    //获取tcp层的数据报文
    char *payload = (char*)(data + sizeof(*eth) + ip->ihl*4 + tcp->doff*4);

    struct data_ip *ipdata;
    ipdata=bpf_ringbuf_reserve(&ip_map, sizeof(*ipdata), 0); //在ringbuf中预留缓冲区大小
    if(!ipdata){
      return 0;
    }
    ipdata->sip=bpf_ntohl(ip->saddr); //将网络字节序转换为主机字节序 32位
    ipdata->dip=bpf_ntohl(ip->daddr);
    ipdata->sport=bpf_ntohl(tcp->source); //16位
    ipdata->dport=bpf_ntohl(tcp->dest);
    bpf_probe_read_kernel(ipdata->payload, HTTP_PAYLOAD_MAX, payload); //拷贝报文数据

    bpf_ringbuf_submit(ipdata, 0); //将ringbuf数据发送到用户态

    return XDP_PASS; //放行
}

char _license[] SEC("license") = "GPL";
