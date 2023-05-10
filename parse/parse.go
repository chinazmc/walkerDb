package parse

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

var (
	protocolErr       = errors.New("protocol error")
	CRLF              = []byte("\r\n")
	PING              = []byte("ping")
	DBSIZE            = []byte("dbsize")
	ERROR_UNSUPPORTED = []byte("-ERR unsupported command\r\n")
	OK                = []byte("+OK\r\n")
	InternalErr       = []byte("-ERR internal error\r\n")
	NotLeaderErr      = []byte("-ERR cluster hasnt leader")
	PONG              = []byte("+PONG\r\n")
	GET               = []byte("get")
	SET               = []byte("set")
	SETEX             = []byte("setex")
	DEL               = []byte("del")
	JOIN              = []byte("join")
	LEVELGET          = []byte("levelget")
	NIL               = []byte("$-1\r\n")
	CZERO             = []byte(":0\r\n")
	CONE              = []byte(":1\r\n")
	BulkSign          = []byte("$")
)

type Request struct {
	Args [][]byte
	Buf  *bytes.Buffer
}

func readLine(r *bufio.Reader) ([]byte, error) {
	p, err := r.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return nil, protocolErr
	}
	return p[:i], nil
}
func Btoi(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	i := 0
	sign := 1
	if data[0] == '-' {
		i++
		sign *= -1
	}
	if i >= len(data) {
		return 0, protocolErr
	}
	var l int
	for ; i < len(data); i++ {
		c := data[i]
		if c < '0' || c > '9' {
			return 0, protocolErr
		}
		l = l*10 + int(c-'0')
	}
	return sign * l, nil
}

func lower(data []byte) {
	for i := 0; i < len(data); i++ {
		if data[i] >= 'A' && data[i] <= 'Z' {
			data[i] += 'a' - 'A'
		}
	}
}
func copyN(buffer *bytes.Buffer, r *bufio.Reader, n int64) (err error) {
	if n <= 512 {
		var buf [512]byte
		_, err = r.Read(buf[:n])
		if err != nil {
			return
		}
		buffer.Write(buf[:n])
	} else {
		_, err = io.CopyN(buffer, r, n)
	}
	return
}
func ReadClient(r *bufio.Reader, req *Request) (err error) {
	line, err := readLine(r)
	if err != nil {
		return
	}
	if len(line) == 0 || line[0] != '*' {
		err = protocolErr
		return
	}
	argc, err := Btoi(line[1:])
	if err != nil {
		return
	}
	if argc <= 0 || argc > 4 {
		err = protocolErr
		return
	}
	var argStarts [4]int
	var argEnds [4]int
	req.Buf.Write(line)
	req.Buf.Write(CRLF)
	cursor := len(line) + 2
	for i := 0; i < argc; i++ {
		line, err = readLine(r)
		if err != nil {
			return
		}
		if len(line) == 0 || line[0] != '$' {
			err = protocolErr
			return
		}
		var argLen int
		argLen, err = Btoi(line[1:])
		if err != nil {
			return
		}
		if argLen < 0 || argLen > 512*1024*1024 {
			err = protocolErr
			return
		}
		req.Buf.Write(line)
		req.Buf.Write(CRLF)
		cursor += len(line) + 2
		err = copyN(req.Buf, r, int64(argLen)+2)
		if err != nil {
			return
		}
		argStarts[i] = cursor
		argEnds[i] = cursor + argLen
		cursor += argLen + 2
	}
	data := req.Buf.Bytes()
	for i := 0; i < argc; i++ {
		req.Args = append(req.Args, data[argStarts[i]:argEnds[i]])
	}
	lower(req.Args[0])
	return
}
