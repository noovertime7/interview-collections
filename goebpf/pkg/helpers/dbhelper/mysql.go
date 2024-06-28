package dbhelper

import (
	"bytes"
	"fmt"
)

func ExtractSQLFromMySQLPacket(packet []byte) (string, error) {
	// MySQL 查询请求报文中的第 5 个字节开始即为 SQL 语句
	sqlStartIndex := 4

	// 校验报文是否符合 MySQL 协议
	if len(packet) < sqlStartIndex+1 {
		return "", fmt.Errorf("Invalid MySQL packet")
	}

	// mysql报文前3个字节为报文长度，第4个字节为序列号。获取方法从mysql/driver看的
	packetLength := int(uint32(packet[0]) | uint32(packet[1])<<8 | uint32(packet[2])<<16)

	// 提取 SQL 语句
	sqlBytes := bytes.TrimRight(packet[sqlStartIndex:], "\x00")

	// 校验报文长度是否正确
	if len(sqlBytes) != packetLength {
		return "", fmt.Errorf("Invalid MySQL packet length")
	}

	//todo 可以从报文的首个字节获取sql的类型，进行进一步解析，比如查询、插入、更新等
	//具体方法查阅driver/mysql中是如何拼接报文的

	return string(sqlBytes), nil
}
