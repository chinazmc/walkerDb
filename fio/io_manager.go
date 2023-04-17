package fio

const DataFilePerm = 0644

// IOManager抽象IO管理接口，可以接入不同的IO类型，目前支持标准文件io
type IOManager interface {
	//从文件的给定位置读取对应的数据
	Read([]byte, int64) (int, error)

	Write([]byte) (int, error)
	Sync() error
	Close() error
	// Size 获取到文件大小
	Size() (int64, error)
}

// NewIOManager 初始化 IOManager，目前只支持标准 FileIO
func NewIOManager(fileName string) (IOManager, error) {
	return NewFileIOManager(fileName)
}
