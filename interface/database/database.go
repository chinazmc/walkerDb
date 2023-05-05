package databaseface

import (
	"walkerDb/connection"
	"walkerDb/reply"
)

// CmdLine 是[][]字节的别名，表示命令行
type CmdLine = [][]byte

// Database redis风格存储接口
type Database interface {
	Exec(client *connection.Connection, args [][]byte) reply.Reply
	AfterClientClose(c connection.Connection)
	Close()
}

// DataEntity 存储绑定到键的数据，包括字符串、列表、哈希、集等
type DataEntity struct {
	Data interface{}
}
