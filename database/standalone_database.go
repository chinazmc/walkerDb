package database

import (
	"fmt"
	"runtime/debug"
	"strings"
	"walkerDb/connection"
	"walkerDb/logger"
	"walkerDb/reply"
)

// StandaloneDatabase
type StandaloneDatabase struct {
	db *DB
}

func NewStandaloneDatabase() *StandaloneDatabase {
	opts := DefaultOptions
	opts.DirPath = "/tmp/bitcask-go"
	db, err := Open(opts)
	if err != nil {
		panic(err)
	}
	return &StandaloneDatabase{db: db}
}

// Exec 执行command
func (mdb *StandaloneDatabase) Exec(c *connection.Connection, cmdLine [][]byte) (result reply.Reply) {

	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()

	cmdName := strings.ToLower(string(cmdLine[0]))
	switch cmdName {
	case "set":
		err := mdb.db.Put(cmdLine[1], cmdLine[2])
		if err != nil {
			return reply.MakeErrReply(err.Error())
		}
		return &reply.NullBulkReply{}
	case "get":
		res, err := mdb.db.Get(cmdLine[1])
		if err != nil {
			return reply.MakeErrReply(err.Error())
		}
		if res == nil {
			return &reply.NullBulkReply{}
		}
		return reply.MakeBulkReply(res)

	case "del":
		err := mdb.db.Delete(cmdLine[1])
		if err != nil {
			return reply.MakeErrReply(err.Error())
		}
		return reply.MakeIntReply(1)
	default:
		return reply.MakeErrReply("error operator type")
	}
}

// Close 优雅关机
func (mdb *StandaloneDatabase) Close() {

}

func (mdb *StandaloneDatabase) AfterClientClose(c connection.Connection) {
}
