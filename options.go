package walkerDb

type Options struct {
	// 数据库数据目录
	DirPath string

	// 数据文件的大小
	DataFileSize int64

	// 每次写数据是否持久化
	SyncWrites bool
	//索引类型
	IndexType IndexerType
}
type IndexerType = int8

const (
	//Btree 索引
	Btree IndexerType = iota + 1
	//ART自适应基数树索引
	ART
)
