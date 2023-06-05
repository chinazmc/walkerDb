package benchmark

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
	"walkerDb/database"
	databaseface "walkerDb/interface/database"
	"walkerDb/utils"
)

var db *database.StandaloneDatabase

func init() {
	//初始化用于基准测试的存储引擎
	options := databaseface.StandaloneDatabaseConfig{
		DataDir: "./tmp/bitcask-go",
	}
	db = database.NewStandaloneDatabase(options)

}
func Benchmark_Put(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := db.Set(utils.GetTestKey(i), utils.RandomValue(1024))
		assert.Nil(b, err)
	}
}

func Benchmark_Get(b *testing.B) {
	for i := 0; i < 10000; i++ {
		err := db.Set(utils.GetTestKey(i), utils.RandomValue(1024))
		assert.Nil(b, err)
	}

	rand.Seed(time.Now().UnixNano())
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := db.Get(utils.GetTestKey(rand.Int()))
		if err != nil && err != database.ErrKeyNotFound {
			b.Fatal(err)
		}
	}
}

func Benchmark_Delete(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < b.N; i++ {
		err, _ := db.Del(utils.GetTestKey(rand.Int()))
		assert.Nil(b, err)
	}
}
