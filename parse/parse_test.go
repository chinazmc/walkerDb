package parse

import (
	"bytes"
	"io"
	"testing"
	"walkerDb/reply"
)

func TestMultiBulkReply(t *testing.T) {
	req := reply.MakeMultiBulkReply([][]byte{
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
	reader := bytes.NewReader(req.ToBytes())
	ch := ParseStream(reader)
	for result := range ch {
		if result.Err != nil {
			if result.Err == io.EOF {
				continue
			}
			t.Error(result.Err)
			continue
		}
		t.Logf("reslut:%s", string(result.Data.ToBytes()))
	}
}
func TestBulkReply(t *testing.T) {
	req := reply.MakeBulkReply([]byte("a\r\nb"))
	reader := bytes.NewReader(req.ToBytes())
	ch := ParseStream(reader)
	for result := range ch {
		if result.Err != nil {
			if result.Err == io.EOF {
				continue
			}
			t.Error(result.Err)
			continue
		}
		t.Logf("reslut:%s", string(result.Data.ToBytes()))
	}
}
