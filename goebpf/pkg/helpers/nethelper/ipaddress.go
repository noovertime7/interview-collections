package nethelper

import (
	"encoding/binary"
	"net"
)

// 将u32的网络ip地址转成net.IP
// 输入的ip地址必须是一个主机字节序的u32
func ResolveIP(input_ip uint32, isbig bool) net.IP {
	ipNetworkOrder := make([]byte, 4)
	if isbig {
		binary.BigEndian.PutUint32(ipNetworkOrder, input_ip)
	} else {
		binary.LittleEndian.PutUint32(ipNetworkOrder, input_ip)
	}

	return ipNetworkOrder
}
