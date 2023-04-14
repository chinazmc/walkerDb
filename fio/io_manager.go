package fio

const DataFilePerm = 0644

// IOManager抽象IO管理接口，可以接入不同的IO类型，目前支持标准文件io
type IOManager interface {
	//从文件的给定位置读取对应的数据
	Read([]byte, int64) (int, error)

	Write([]byte) (int, error)
	Sync() error
	Close() error
}
