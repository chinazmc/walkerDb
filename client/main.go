package main

import (
	"fmt"
	"net"
	"walkerDb/reply"
)

// 客户端
func main() {
	conn, err := net.Dial("tcp", "0.0.0.0:8972")
	if err != nil {
		fmt.Println("err :", err)
		return
	}
	defer conn.Close() // 关闭连接

	req := reply.MakeMultiBulkReply([][]byte{
		[]byte("set"),
		[]byte("my"),
		[]byte("first"),
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
	//---------------------------------------------
	req = reply.MakeMultiBulkReply([][]byte{
		[]byte("set"),
		[]byte("my"),
		[]byte("second"),
	})
	_, err = conn.Write(req.ToBytes()) // 发送数据
	if err != nil {
		return
	}
	buf = [512]byte{}
	n, err = conn.Read(buf[:])
	if err != nil {
		fmt.Println("recv failed, err:", err)
		return
	}
	fmt.Println(string(buf[:n]))
	//---------------------------------------------
	req = reply.MakeMultiBulkReply([][]byte{
		[]byte("get"),
		[]byte("my"),
		//[]byte("second"),
	})
	_, err = conn.Write(req.ToBytes()) // 发送数据
	if err != nil {
		return
	}
	buf = [512]byte{}
	n, err = conn.Read(buf[:])
	if err != nil {
		fmt.Println("recv failed, err:", err)
		return
	}
	fmt.Println(string(buf[:n]))
}
