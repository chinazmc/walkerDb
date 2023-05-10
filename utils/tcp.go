package utils

import (
	"net"
)

func SendTcpReq(address string, req []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	defer conn.Close() // 关闭连接

	_, err = conn.Write(req) // 发送数据
	if err != nil {
		return nil, err
	}
	buf := [1024]byte{}
	n, err := conn.Read(buf[:])
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
