package nethelper

import (
	"github.com/vishvananda/netlink"
	"log"
	"net"
)

// 获取所有的veth网卡
func GetVeths() []*net.Interface {
	links, err := netlink.LinkList()
	if err != nil {
		log.Fatal(err)
	}
	var veths []*net.Interface
	for _, link := range links {
		if link.Type() == "veth" {
			iface, err := net.InterfaceByName(link.Attrs().Name)
			if err != nil {
				log.Fatal(err)
			}
			veths = append(veths, iface)
		}
	}
	return veths
}
