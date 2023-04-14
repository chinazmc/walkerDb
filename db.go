package walkerDb

import (
	"sync"
	"walkerDb/data"
	"walkerDb/index"
)

// DB 存储引擎实例
type DB struct {
	options    Options
	mu         *sync.RWMutex
	activeFile *data.DataFile
	olderFiles map[uint32]*data.DataFile
	index      index.Indexer
}

// Put 写入Key/value数据，key不能为空
func (db *DB) Put(key []byte, value []byte) error {
	if len(key) == 0 {
		return ErrKeyIsEmpty
	}
	logRecord := &data.LogRecord{
		Key:   key,
		Value: value,
		Type:  data.LogRecordNormal,
	}
	//追加写入到当前活跃数据文件当中
	pos, err := db.appendLogRecord(logRecord)
	if err != nil {
		return err
	}

	// 更新内存索引
	if ok := db.index.Put(key, pos); !ok {
		return ErrIndexUpdateFailed
	}

	return nil
}

// Get 根据 key 读取数据
func (db *DB) Get(key []byte) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// 判断 key 的有效性
	if len(key) == 0 {
		return nil, ErrKeyIsEmpty
	}

	// 从内存数据结构中取出 key 对应的索引信息
	logRecordPos := db.index.Get(key)
	// 如果 key 不在内存索引中，说明 key 不存在
	if logRecordPos == nil {
		return nil, ErrKeyNotFound
	}

	// 根据文件 id 找到对应的数据文件
	var dataFile *data.DataFile
	if db.activeFile.FileId == logRecordPos.Fid {
		dataFile = db.activeFile
	} else {
		dataFile = db.olderFiles[logRecordPos.Fid]
	}
	// 数据文件为空
	if dataFile == nil {
		return nil, ErrDataFileNotFound
	}

	// 根据偏移读取对应的数据
	logRecord, err := dataFile.ReadLogRecord(logRecordPos.Offset)
	if err != nil {
		return nil, err
	}

	if logRecord.Type == data.LogRecordDeleted {
		return nil, ErrKeyNotFound
	}

	return logRecord.Value, nil
}
func (db *DB) appendLogRecord(logRecord *data.LogRecord) (*data.LogRecordPos, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	//判断当前活跃数据文件是否存在，因为数据库在没有写入的时候是没有文件生成的
	//如果为空则初始化为数据文件
	if db.activeFile == nil {
		if err := db.setActiveDataFile(); err != nil {
			return nil, err
		}
	}
	//写入数据编码
	encRecord, size := data.EncodeLogRecord(logRecord)
	//如果写入的数据已经到达活跃文件的阈值，就关闭活跃文件并打开新文件
	if db.activeFile.WriteOff+size > db.options.DataFileSize {
		//先持久化文件
		if err := db.activeFile.Sync(); err != nil {
			return nil, err
		}
		db.olderFiles[db.activeFile.FileId] = db.activeFile
		if err := db.setActiveDataFile(); err != nil {
			return nil, err
		}
	}
	writeOff := db.activeFile.WriteOff
	if err := db.activeFile.Write(encRecord); err != nil {
		return nil, err
	}
	if db.options.SyncWrites {
		if err := db.activeFile.Sync(); err != nil {
			return nil, err
		}
	}
	//构造内存索引信息
	pos := &data.LogRecordPos{Fid: db.activeFile.FileId, Offset: writeOff}
	return pos, nil
}

// 设置当前活跃文件
// 在访问此方法之前必须持有互斥锁
func (db *DB) setActiveDataFile() error {
	var initialFileId uint32 = 0
	if db.activeFile == nil {
		initialFileId = db.activeFile.FileId + 1
	}
	// 打开新的数据文件
	dataFile, err := data.OpenDataFile(db.options.DirPath, initialFileId)
	if err != nil {
		return err
	}
	db.activeFile = dataFile
	return nil
}
