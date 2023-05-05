package main

import (
	"fmt"
	"net"
)

// 客户端
func main() {
	conn, err := net.Dial("tcp", "0.0.0.0:8972")
	if err != nil {
		fmt.Println("err :", err)
		return
	}
	defer conn.Close()                         // 关闭连接
	_, err = conn.Write([]byte("hello world")) // 发送数据
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
	_, err = conn.Write([]byte("hello zmc")) // 发送数据
	if err != nil {
		return
	}
	buf2 := [512]byte{}
	n, err = conn.Read(buf2[:])
	if err != nil {
		fmt.Println("recv failed, err:", err)
		return
	}
	fmt.Println(string(buf2[:n]))
}
