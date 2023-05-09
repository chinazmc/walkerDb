package database

import (
	databaseface "walkerDb/interface/database"
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
func (mdb *StandaloneDatabase) Exec(cmdLine [][]byte) (result []byte) {
	return
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
func (mdb *StandaloneDatabase) Del(key []byte) (error, bool) {
	err := mdb.db.Delete(key)
	if err != nil {
		return err, false
	}
	return nil, true
}
func (mdb *StandaloneDatabase) SetEX(key []byte, value []byte, expireSeconds int) error {
	//todo
	return nil
}
