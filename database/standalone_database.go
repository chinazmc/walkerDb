package database

import (
	"errors"
	"time"
	databaseface "walkerDb/interface/database"
)

var (
	ErrWrongTypeOperation = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
)

type redisDataType = byte

const (
	String redisDataType = iota
	Hash
	Set
	List
	ZSet
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
func (mdb *StandaloneDatabase) Start() {

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
func (rds *StandaloneDatabase) Type(key []byte) (redisDataType, error) {
	encValue, err := rds.db.Get(key)
	if err != nil {
		return 0, err
	}
	if len(encValue) == 0 {
		return 0, errors.New("value is null")
	}
	// 第一个字节就是类型
	return encValue[0], nil
}

// ============string 数据结构===================
//
//	func (rds *StandaloneDatabase) Set(key []byte, ttl time.Duration, value []byte) error {
//		if value == nil {
//			return nil
//		}
//		//编码value:type+expire+payload
//		buf := make([]byte, binary.MaxVarintLen64+1)
//		buf[0] = String
//		var index = 1
//		var expire int64 = 0
//		if ttl != 0 {
//			expire = time.Now().Add(ttl).UnixNano()
//		}
//		index += binary.PutVarint(buf[index:], expire)
//		encValue := make([]byte, index+len(value))
//		copy(encValue[:index], buf[:index])
//		copy(encValue[index:], value)
//		return rds.db.Put(key, encValue)
//	}
//
//	func (rds *StandaloneDatabase) Get(key []byte) ([]byte, error) {
//		encValue, err := rds.db.Get(key)
//		if err != nil {
//			return nil, err
//		}
//
//		// 解码
//		dataType := encValue[0]
//		if dataType != String {
//			return nil, ErrWrongTypeOperation
//		}
//
//		var index = 1
//		expire, n := binary.Varint(encValue[index:])
//		index += n
//		// 判断是否过期
//		if expire > 0 && expire <= time.Now().UnixNano() {
//			return nil, nil
//		}
//
//		return encValue[index:], nil
//	}
//
// ======================= Hash 数据结构 =======================
func (rds *StandaloneDatabase) HSet(key, field, value []byte) (bool, error) {
	//先查找元数据
	meta, err := rds.findMetadata(key, Hash)
	if err != nil {
		return false, err
	}
	// 构造 Hash 数据部分的 key
	hk := &hashInternalKey{
		key:     key,
		version: meta.version,
		field:   field,
	}
	encKey := hk.encode()
	//先查找是否存在
	var exist = true
	if _, err = rds.db.Get(encKey); err == ErrKeyNotFound {
		exist = false
	}
	wb := rds.db.NewWriteBatch(DefaultWriteBatchOptions)
	//不存在则更新元数据
	if !exist {
		meta.size++
		_ = wb.Put(key, meta.encode())
	}
	_ = wb.Put(encKey, value)
	if err = wb.Commit(); err != nil {
		return false, err
	}
	return !exist, nil
}
func (rds *StandaloneDatabase) HGet(key, field []byte) ([]byte, error) {
	meta, err := rds.findMetadata(key, Hash)
	if err != nil {
		return nil, err
	}
	if meta.size == 0 {
		return nil, nil
	}

	hk := &hashInternalKey{
		key:     key,
		version: meta.version,
		field:   field,
	}

	return rds.db.Get(hk.encode())
}

func (rds *StandaloneDatabase) HDel(key, field []byte) (bool, error) {
	meta, err := rds.findMetadata(key, Hash)
	if err != nil {
		return false, err
	}
	if meta.size == 0 {
		return false, nil
	}

	hk := &hashInternalKey{
		key:     key,
		version: meta.version,
		field:   field,
	}
	encKey := hk.encode()

	// 先查看是否存在
	var exist = true
	if _, err = rds.db.Get(encKey); err == ErrKeyNotFound {
		exist = false
	}

	if exist {
		wb := rds.db.NewWriteBatch(DefaultWriteBatchOptions)
		meta.size--
		_ = wb.Put(key, meta.encode())
		_ = wb.Delete(encKey)
		if err = wb.Commit(); err != nil {
			return false, err
		}
	}

	return exist, nil
}
func (rds *StandaloneDatabase) findMetadata(key []byte, dataType redisDataType) (*metadata, error) {
	metaBuf, err := rds.db.Get(key)
	if err != nil && err != ErrKeyNotFound {
		return nil, err
	}
	var meta *metadata
	var exist = true
	if err == ErrKeyNotFound {
		exist = false
	} else {
		meta = decodeMetadata(metaBuf)
		// 判断数据类型
		if meta.dataType != dataType {
			return nil, ErrWrongTypeOperation
		}
		// 判断过期时间
		if meta.expire != 0 && meta.expire <= time.Now().UnixNano() {
			exist = false
		}
	}

	if !exist {
		meta = &metadata{
			dataType: dataType,
			expire:   0,
			version:  time.Now().UnixNano(),
			size:     0,
		}
		if dataType == List {
			meta.head = initialListMark
			meta.tail = initialListMark
		}
	}
	return meta, nil
}
