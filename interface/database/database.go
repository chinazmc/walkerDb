package databaseface

// CmdLine 是[][]字节的别名，表示命令行
type CmdLine = [][]byte

// Database redis风格存储接口
type Database interface {
	//Exec(args [][]byte) reply.Reply
	Exec(cmdLine [][]byte) (result []byte)
	//AfterClientClose(c connection.Connection)
	Close()
	Set(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	// Del deletes an item in the cache by key and returns true or false if a delete occurred.
	Del(key []byte) (error, bool)
	SetEX(key []byte, value []byte, expireSeconds int) error
}

// DataEntity 存储绑定到键的数据，包括字符串、列表、哈希、集等
type DataEntity struct {
	Data interface{}
}
