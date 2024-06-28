package dbhelper

import (
	"bytes"
	"errors"
	"github.com/tidwall/redcon"
	"strings"
)

// 解析redis报文
func ExtractSQLFromRedisPacket(packet []byte) (string, error) {
	//去除报文中的空字节
	packet = bytes.TrimRight(packet, "\x00")
	n, rsp := redcon.ReadNextRESP(packet)

	var ret []string
	if !isEmptyRESP(rsp) && n == len(packet) {
		rsp.ForEach(func(resp redcon.RESP) bool {
			ret = append(ret, resp.String())
			return true
		})
		//重新拼接，原rsp是这样的$3\r\nset\r\n$4\r\nname\r\n$8\r\nzhangsan
		return strings.Join(ret, " "), nil
	}

	return "", errors.New("Invalid Redis packet")
}

// 源码来的
func isEmptyRESP(resp redcon.RESP) bool {
	return resp.Type == 0 && resp.Count == 0 &&
		resp.Data == nil && resp.Raw == nil
}
