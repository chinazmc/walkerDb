package tcp

import (
	"bytes"
	"flag"
	"strconv"
	"sync"
	"sync/atomic"
	. "walkerDb/parse"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)
var Epoller *epoll
var WorkerPool *pool
var (
	c = flag.Int("c", 10, "concurrency")
)

type pool struct {
	workers   int
	maxTasks  int
	taskQueue chan *ConnArgs

	mu     sync.Mutex
	closed atomic.Bool
	done   chan struct{}
}

func newPool(w int, t int) *pool {
	return &pool{
		workers:   w,
		maxTasks:  t,
		taskQueue: make(chan *ConnArgs, t),
		done:      make(chan struct{}),
	}
}

func (p *pool) Close() {
	p.mu.Lock()
	p.closed.Store(true)
	close(p.done)
	close(p.taskQueue)
	p.mu.Unlock()
}

func (p *pool) addTask(connArg *ConnArgs) {
	p.mu.Lock()
	if p.closed.Load() {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	p.taskQueue <- connArg
}

func (p *pool) start() {
	for i := 0; i < p.workers; i++ {
		go p.startWorker()
	}
}

func (p *pool) startWorker() {
	for {
		select {
		case <-p.done:
			return
		case connArgs := <-p.taskQueue:
			if connArgs != nil {
				p.handleConn(connArgs)
			}
		}
	}
}

//	func (h *pool) closeClient(client *connection.Connection) {
//		if err := Epoller.Remove(client.GetConn()); err != nil {
//			logger.Info("failed to remove %v", err)
//		}
//		_ = client.Close()
//		h.activeConn.Delete(client)
//	}
func (p *pool) handleConn(connArgs *ConnArgs) {
	if p.closed.Load() {
		//如果已经关闭了，就不要处理这个链接了
		_ = connArgs.conn.Close()
		return
	}
	req := connArgs.req
	reply := new(bytes.Buffer)
	cache := connArgs.conn.server.cache
	if len(req.Args) == 4 {
		if bytes.Equal(req.Args[0], SETEX) {
			expire, err := Btoi(req.Args[2])
			if err != nil {
				reply.Write(ERROR_UNSUPPORTED)
			} else {
				cache.SetEX(req.Args[1], req.Args[3], expire)
				reply.Write(OK)
			}
		}
		if bytes.Equal(req.Args[0], JOIN) {
			res := cache.Exec(req.Args)
			reply.Write(res)
		}
	} else if len(req.Args) == 3 {
		if bytes.Equal(req.Args[0], SET) {
			err := cache.Set(req.Args[1], req.Args[2])
			if err != nil {
				reply.Write([]byte(err.Error()))
			} else {
				reply.Write(OK)
			}
		}
		if bytes.Equal(req.Args[0], LEVELGET) {
			res := cache.Exec(req.Args)
			reply.Write(res)
		}
	} else if len(req.Args) == 2 {
		if bytes.Equal(req.Args[0], GET) {
			value, err := cache.Get(req.Args[1])
			if err != nil {
				reply.Write(NIL)
			} else {
				bukLen := strconv.Itoa(len(value))
				reply.Write(BulkSign)
				reply.WriteString(bukLen)
				reply.Write(CRLF)
				reply.Write(value)
				reply.Write(CRLF)
			}
		} else if bytes.Equal(req.Args[0], DEL) {
			if _, isDel := cache.Del(req.Args[1]); isDel {
				reply.Write(CONE)
			} else {
				reply.Write(CZERO)
			}
		}
	} else if len(req.Args) == 1 {
		if bytes.Equal(req.Args[0], PING) {
			reply.Write(PONG)
		} else {
			reply.Write(ERROR_UNSUPPORTED)
		}
	}
	_, err := connArgs.conn.conn.Write(reply.Bytes())
	if err != nil {
		connArgs.conn.conn.Close()
		return
	}
}
