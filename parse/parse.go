package parse

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"runtime/debug"
	"strconv"
)

type OperateType = byte

const (
	OperateGetType OperateType = iota
	OperatePutType
	OperateDeleteType
)

type Payload struct {
	Key   []byte
	Value []byte
	Type  OperateType
}

//用这个协议来解决粘包
// RESP 通过第一个字符来表示格式：
//
// 简单字符串：以"+" 开始， 如："+OK\r\n"
// 错误：以"-" 开始，如："-ERR Invalid Synatx\r\n"
// 整数：以":"开始，如：":1\r\n"
// 字符串：以 $ 开始
// 数组：以 * 开始
// 参考https://www.cnblogs.com/Finley/p/11923168.html

// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// ParseOne reads data from []byte and return the first payload
func ParseOne(data []byte) ([]*Payload, error) {
	reader := bytes.NewReader(data)
	payload, err := parse0(reader)
	if payload == nil {
		return nil, errors.New("no protocol")
	}
	return payload, err
}

func parse0(rawReader io.Reader) ([]*Payload, error) {
	defer func() {
		if err := recover(); err != nil {
			//logger.Error(err, string(debug.Stack())) //todo
			println(string(debug.Stack()))
		}
	}()
	reader := bufio.NewReader(rawReader)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		length := len(line)
		if length <= 2 || line[length-2] != '\r' {
			// there are some empty lines within replication traffic, ignore this error
			//protocolError(ch, "empty line")
			continue
		}
		line = bytes.TrimSuffix(line, []byte{'\r', '\n'})
		switch line[0] {
		case '+':
		case '-':
		case ':':
		case '$':
		case '*':
			data, err := parseArray(line, reader)
			if err != nil {
				return nil, err
			}
			return data, nil
		default:
			return nil, errors.New("error!!!!!!!!!")
		}
	}
}

func parseArray(header []byte, reader *bufio.Reader) ([]*Payload, error) {
	nStrs, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil || nStrs < 0 {
		err := protocolError("illegal array header " + string(header[1:]))
		return nil, err
	} else if nStrs == 0 {
		return nil, nil
	}
	lines := make([][]byte, 0, nStrs)
	for i := int64(0); i < nStrs; i++ {
		var line []byte
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		length := len(line)
		if length < 4 || line[length-2] != '\r' || line[0] != '$' {
			protocolError("illegal bulk string header " + string(line))
			break
		}
		strLen, err := strconv.ParseInt(string(line[1:length-2]), 10, 64)
		if err != nil || strLen < -1 {
			protocolError("illegal bulk string length " + string(line))
			break
		}
		if strLen == -1 {
			lines = append(lines, []byte{})
			continue
		}
		body := make([]byte, strLen+2)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			return nil, err
		}
		lines = append(lines, body[:len(body)-2])
	}
	return getPayLoadArr(lines), nil
}
func getPayLoadArr(lines [][]byte) []*Payload {
	arr := []*Payload{}
	for i := 0; i < len(lines); {
		switch string(lines[i]) {
		case "set":
			if i+2 <= len(lines) {
				arr = append(arr, &Payload{
					Type:  OperatePutType,
					Key:   lines[i+1],
					Value: lines[i+2],
				})
				i = i + 3
			}
		case "get":
			if i+1 <= len(lines) {
				arr = append(arr, &Payload{
					Type: OperateGetType,
					Key:  lines[i+1],
				})
				i = i + 2
			}
		case "del":
			if i+1 <= len(lines) {
				arr = append(arr, &Payload{
					Type: OperateDeleteType,
					Key:  lines[i+1],
				})
				i = i + 2
			}
		default:
			i++
		}
	}
	return arr
}

func protocolError(msg string) error {
	return errors.New("protocol error: " + msg)
}

var (

	// CRLF is the line separator of redis serialization protocol
	CRLF = "\r\n"
)

// MultiBulkReply stores a list of string
type MultiBulkReply struct {
	Args [][]byte
}

// MakeMultiBulkReply creates MultiBulkReply
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

// ToBytes marshal redis.Reply
func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args {
		if arg == nil {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}
