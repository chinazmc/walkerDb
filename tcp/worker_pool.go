package tcp

import (
	"bytes"
	"flag"
	"strconv"
	"sync"
	"sync/atomic"
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
	if len(req.args) == 4 {
		if bytes.Equal(req.args[0], SETEX) {
			expire, err := btoi(req.args[2])
			if err != nil {
				reply.Write(ERROR_UNSUPPORTED)
			} else {
				cache.SetEX(req.args[1], req.args[3], expire)
				reply.Write(OK)
			}
		}
		if bytes.Equal(req.args[0], JOIN) {
			res := cache.Exec(req.args)
			reply.Write(res)
		}
	} else if len(req.args) == 3 {
		if bytes.Equal(req.args[0], SET) {
			cache.Set(req.args[1], req.args[2])
			reply.Write(OK)
		}
		if bytes.Equal(req.args[0], LEVELGET) {
			cache.Exec(req.args)
		}
	} else if len(req.args) == 2 {
		if bytes.Equal(req.args[0], GET) {
			value, err := cache.Get(req.args[1])
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
		} else if bytes.Equal(req.args[0], DEL) {
			if _, isDel := cache.Del(req.args[1]); isDel {
				reply.Write(CONE)
			} else {
				reply.Write(CZERO)
			}
		}
	} else if len(req.args) == 1 {
		if bytes.Equal(req.args[0], PING) {
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
	//ch := parse.ParseStream(conn)
	//for payload := range ch {
	//	if payload.Err != nil {
	//		if payload.Err == io.EOF ||
	//			payload.Err == io.ErrUnexpectedEOF ||
	//			strings.Contains(payload.Err.Error(), "use of closed network connection") {
	//			// connection closed
	//			p.closeClient(client)
	//			logger.Error("connection closed: " + client.RemoteAddr().String())
	//			return
	//		}
	//		// protocol err
	//		errReply := reply.MakeErrReply(payload.Err.Error())
	//		err := client.Write(errReply.ToBytes())
	//		if err != nil {
	//			p.closeClient(client)
	//			logger.Info("connection closed: " + client.RemoteAddr().String())
	//			return
	//		}
	//		continue
	//	}
	//	if payload.Data == nil {
	//		logger.Error("empty payload")
	//		continue
	//	}
	//	r, ok := payload.Data.(*reply.MultiBulkReply)
	//	if !ok {
	//		logger.Error("require multi bulk reply")
	//		continue
	//	}
	//	result := p.db.Exec(client, r.Args)
	//	if result != nil {
	//		_ = client.Write(result.ToBytes())
	//	} else {
	//		_ = client.Write(unknownErrReplyBytes)
	//	}
	//}

}
