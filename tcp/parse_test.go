package tcp

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"walkerDb/reply"
)

func TestMultiBulkReply(t *testing.T) {
	mockReq := reply.MakeMultiBulkReply([][]byte{
		[]byte("set"),
		[]byte("hello"),
		[]byte("world"),
	})
	reader := bufio.NewReader(strings.NewReader(string(mockReq.ToBytes())))
	var req = new(Request)
	req.buf = new(bytes.Buffer)
	err := ReadClient(reader, req)
	if err != nil {
		t.Error(err)
	} else {
		for _, item := range req.args {
			t.Log(string(item))
		}
	}
}
