package index

import (
	"bytes"
	"github.com/google/btree"
	"walkerDb/data"
)

// Indexer抽象索引接口，后续如果想要接入其他的数据结构，直接实现这个接口
type Indexer interface {
	//Put 向索引中存储key对应的数据位置信息
	Put(key []byte, pos *data.LogRecordPos) bool
	//Get根据key取出对应的索引位置信息
	Get(key []byte) *data.LogRecordPos
	//Delete 根据key删除对应的索引位置信息
	Delete(key []byte) bool
}
type Item struct {
	key []byte
	pos *data.LogRecordPos
}

func (ai *Item) Less(bi btree.Item) bool {
	//0是等于，-1是小于，1是大于
	return bytes.Compare(ai.key, bi.(*Item).key) == -1
}
