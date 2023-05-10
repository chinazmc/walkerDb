package main

import (
	"fmt"
	"net"
	"walkerDb/reply"
)

const (
	Default    = "0"
	Stale      = "1"
	Consistent = "2"
)

// 客户端
func main() {
	conn, err := net.Dial("tcp", "0.0.0.0:8033")
	if err != nil {
		fmt.Println("err :", err)
		return
	}
	defer conn.Close()
	req := reply.MakeMultiBulkReply([][]byte{
		[]byte("levelget"),
		[]byte("my"),
		[]byte(Stale),
	})
	_, err = conn.Write(req.ToBytes()) // 发送数据
	if err != nil {
		return
	}
	buf := [512]byte{}
	n, err := conn.Read(buf[:])
	if err != nil {
		fmt.Println("recv failed, err:", err)
		return
	}
	fmt.Println(string(buf[:n]))
}
