//go:build ignore
#include <vmlinux.h>
#include <bpf_helpers.h>
#include <bpf_endian.h>
#include <bpf_tracing.h>
#include <bpf_legacy.h>

#define ETH_HLEN 14 //以太网头部长度
#define IP_CSUM_OFF (ETH_HLEN + offsetof(struct iphdr, check))
#define TOS_OFF (ETH_HLEN + offsetof(struct iphdr, tos))
#define TCP_CSUM_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, check)) //csum的偏移量
#define IP_SRC_OFF (ETH_HLEN + offsetof(struct iphdr, saddr))
#define TCP_DPORT_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, dest)) //目标端口的偏移量
#define TCP_SPORT_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, source)) //目标端口的偏移量

#define IS_PSEUDO 0x10

char LICENSE[] SEC("license") = "GPL";

struct tc_data_ip {
    __u32 sip; //源IP地址
    __u32 dip; //目的IP地址
    __u32 sport; //源端口
    __u32 dport; //目的端口
};

//ringbuf
struct { //ringbuf，环形缓冲区，算是一种用户内核交互的优先选择
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries,1<<20); //大概是10M大小
} tc_ip_map SEC(".maps");

//从skb获取ip头部
static inline int iph_dr(struct __sk_buff *skb, struct iphdr *iph) //内连函数，编译时直接展开，减少函数调用开销
{
    int offset = sizeof(struct ethhdr); //计算以太网头部的偏移量
    return bpf_skb_load_bytes(skb, offset, iph, sizeof(*iph));
}

//从skb获取tcp头部
static inline int tcph_dr(struct __sk_buff *skb, struct tcphdr *tcph) //内连函数，编译时直接展开，减少函数调用开销
{
    int offset = sizeof(struct ethhdr) + sizeof(struct iphdr); //计算以太网头部和ip头部的偏移量
    return bpf_skb_load_bytes(skb, offset, tcph, sizeof(*tcph));
}

//改源ip的，没用上，先注释了
//todo 使用目标ip重定向的问题在于，就是old_ip一定得要真实存在才可以，否则连二层arp都通过不了，需要做arp欺骗
//static inline void set_tcp_ip_src(struct __sk_buff *skb, __u32 new_ip)
//{
//	__u32 old_ip = bpf_htonl(load_word(skb, IP_SRC_OFF));
//
//	bpf_l4_csum_replace(skb, TCP_CSUM_OFF, old_ip, new_ip, IS_PSEUDO | sizeof(new_ip));
//	bpf_l3_csum_replace(skb, IP_CSUM_OFF, old_ip, new_ip, sizeof(new_ip));
//	bpf_skb_store_bytes(skb, IP_SRC_OFF, &new_ip, sizeof(new_ip), 0);
//}

static inline void set_tcp_dest_port(struct __sk_buff *skb, __u16 new_port)
{ //源码 —— https://github.com/torvalds/linux/blob/master/samples/bpf/tcbpf1_kern.c
	__u16 old_port = bpf_htons(load_half(skb, TCP_DPORT_OFF));

	bpf_l4_csum_replace(skb, TCP_CSUM_OFF, old_port, new_port, sizeof(new_port)); //1.修改校验和csum
	bpf_skb_store_bytes(skb, TCP_DPORT_OFF, &new_port, sizeof(new_port), 0); //2.重新存储到skb
}

static inline void set_tcp_src_port(struct __sk_buff *skb, __u16 new_port)
{
	__u16 old_port = bpf_htons(load_half(skb, TCP_SPORT_OFF));

	bpf_l4_csum_replace(skb, TCP_CSUM_OFF, old_port, new_port, sizeof(new_port));
	bpf_skb_store_bytes(skb, TCP_SPORT_OFF, &new_port, sizeof(new_port), 0);
}

SEC("classifier") //代表tc的流量分类
int mytc(struct __sk_buff *skb)
{

    struct iphdr ip;
    iph_dr(skb, &ip);
    struct tcphdr tcp;
    tcph_dr(skb, &tcp);

    //打包网络数据
    //如果ip包是tcp协议，才发送数据
    if(ip.protocol != IPPROTO_TCP){
        return 0;
    }

    //作用：将访问到172.17.0.3:8080重定向到172.17.0.3:80
    __u16 watch_port = bpf_ntohs(tcp.dest); //目标端口
    __u32 watch_ip = bpf_ntohl(0xAC110003);  //172.17.0.3
    if (watch_port == 8080 && ip.daddr == watch_ip) {
        set_tcp_dest_port(skb, bpf_htons(80)); //修改目标端口 A -> B 8080 -> 80
        tcph_dr(skb, &tcp); //重新读取skb数据到tcp
    }
    //这次修改的是tcp三次握手中第二次也就是服务端响应的端口，否则客户端接收到的源端口与目标端口不一致，会重置请求
    __u16 src_port = bpf_ntohs(tcp.source); //源端口
    if (src_port == 80 && ip.saddr == watch_ip) {
        set_tcp_src_port(skb, bpf_htons(8080)); //修改源端口 B -> A 80 -> 8080
        tcph_dr(skb, &tcp);
    }

    //作用：将源ip从172.17.0.2修改成10.0.0.1
    // ntohl = network to host long 网络报文中的ip地址是以16进制的网络字节序存储的，需要转换成主机字节序
//    __u32 watch_ip = bpf_ntohl(0xAC110002); //172.17.0.2
//    if (ip.saddr == watch_ip) {
//        __u32 new_ip = bpf_ntohl(0x0A000001); //10.0.0.1
//        set_tcp_ip_src(skb, new_ip); //修改源ip
//        iph_dr(skb, &ip); //重新读取skb数据到ip
//        tcph_dr(skb, &tcp); //重新读取skb数据到tcp
//    }
// todo 这里失败了 ，因为第二次握手服务端回包的时候，目标ip10.0.0.1并不存在，所以会arp失败，因此没有进入bpf程序，也就不能篡改
//    if (ip.daddr == watch_ip) {
//        __u32 new_ip = bpf_ntohl(0xAC110002); //
//        set_tcp_ip_src(skb, new_ip);
//        iph_dr(skb, &ip);
//        tcph_dr(skb, &tcp);
//    }


    struct tc_data_ip *ipdata;
    ipdata=bpf_ringbuf_reserve(&tc_ip_map, sizeof(*ipdata), 0); //在ringbuf中预留缓冲区大小
    if(!ipdata){
      return 0;
    }
    ipdata->sip = bpf_ntohl(ip.saddr); //网络字节序转换为主机字节序 否则转换成xxx.xxx.xxx.xxx后会颠倒
    ipdata->dip = bpf_ntohl(ip.daddr);
    ipdata->sport = bpf_ntohs(tcp.source);
    ipdata->dport = bpf_ntohs(tcp.dest);
    bpf_ringbuf_submit(ipdata, 0); //提交数据

    return 0; //代表放行，是action的一种，混合了action和classifer，分类器类型需要指定成direct-action
}
