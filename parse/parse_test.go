package parse

import "testing"

func TestParseOne(t *testing.T) {
	req := MakeMultiBulkReply([][]byte{
		[]byte("set"),
		[]byte("hello"),
		[]byte("world"),
		[]byte("set"),
		[]byte("hello1"),
		[]byte("world1"),
		[]byte("get"),
		[]byte("hello"),
		[]byte("del"),
		[]byte("hello1"),
	})
	arr, err := ParseOne(req.ToBytes())
	if err != nil {
		t.Error(err)
	}
	for _, result := range arr {
		t.Logf("key:%s,value:%s,type:%d", string(result.Key), string(result.Value), result.Type)
	}
}
