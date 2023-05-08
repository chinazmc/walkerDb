package database

import (
	"fmt"
	"runtime/debug"
	"strings"
	databaseface "walkerDb/interface/database"
	"walkerDb/logger"
	"walkerDb/reply"
)

// StandaloneDatabase
type StandaloneDatabase struct {
	db *DB
}

func NewStandaloneDatabase(config databaseface.StandaloneDatabaseConfig) *StandaloneDatabase {
	opts := DefaultOptions
	if config.DataDir != "" {
		opts.DirPath = config.DataDir
	}
	db, err := Open(opts)
	if err != nil {
		panic(err)
	}
	return &StandaloneDatabase{db: db}
}

// Exec 执行command
func (mdb *StandaloneDatabase) Exec(cmdLine [][]byte) (result reply.Reply) {

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
	mdb.db.Close()
}

// func (mdb *StandaloneDatabase) AfterClientClose(c connection.Connection) {
// }
func (mdb *StandaloneDatabase) Set(key []byte, value []byte) error {
	return mdb.db.Put(key, value)
}
func (mdb *StandaloneDatabase) Get(key []byte) ([]byte, error) {
	return mdb.db.Get(key)
}

// Del deletes an item in the cache by key and returns true or false if a delete occurred.
func (mdb *StandaloneDatabase) Del(key []byte) bool {
	err := mdb.db.Delete(key)
	if err != nil {
		return false
	}
	return true
}
func (mdb *StandaloneDatabase) SetEX(key []byte, value []byte, expireSeconds int) error {
	//todo
	return nil
}
