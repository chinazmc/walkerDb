package walkerDb

import (
	"flag"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"walkerDb/connection"
	"walkerDb/database"
	databaseface "walkerDb/interface/database"
	"walkerDb/logger"
	"walkerDb/parse"
	"walkerDb/reply"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)
var Epoller *epoll
var WorkerPool *pool
var (
	c = flag.Int("c", 10, "concurrency")
)

func StartEpollAndWorkerPool(listener net.Listener) {
	//todo 使用once
	WorkerPool = newPool(*c, 1000000)
	WorkerPool.start()
	var err error
	Epoller, err = MkEpoll(listener)
	if err != nil {
		panic(err)
	}

	go start()
}
func start() {
	for {
		connections, err := Epoller.Wait()
		if err != nil {
			logger.Info(fmt.Sprintf("failed to epoll wait %v", err))
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}

			WorkerPool.addTask(conn)
		}
	}
}

type pool struct {
	workers   int
	maxTasks  int
	taskQueue chan net.Conn

	mu     sync.Mutex
	closed atomic.Bool
	done   chan struct{}

	activeConn sync.Map // *client -> placeholder
	db         databaseface.Database

	ConnWriteTimeout time.Duration
	ReadTimeout      time.Duration
	KeepAliveTimeout time.Duration
}

func newPool(w int, t int) *pool {
	var db databaseface.Database
	config := database.DefaultDataBaseOptions
	if config.Self != "" &&
		len(config.Peers) > 0 {
		//db = cluster.MakeClusterDatabase()
	} else {
		db = database.NewStandaloneDatabase()
	}
	return &pool{
		workers:          w,
		maxTasks:         t,
		taskQueue:        make(chan net.Conn, t),
		done:             make(chan struct{}),
		db:               db,
		KeepAliveTimeout: 5 * time.Minute,
	}
}

func (p *pool) Close() {
	p.mu.Lock()
	p.closed.Store(true)
	p.activeConn.Range(func(key, value any) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	p.db.Close()
	close(p.done)
	close(p.taskQueue)
	p.mu.Unlock()
}

func (p *pool) addTask(conn net.Conn) {
	p.mu.Lock()
	if p.closed.Load() {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	p.taskQueue <- conn
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
		case conn := <-p.taskQueue:
			if conn != nil {
				p.handleConn(conn)
			}
		}
	}
}
func (h *pool) closeClient(client *connection.Connection) {
	if err := Epoller.Remove(client.GetConn()); err != nil {
		logger.Info("failed to remove %v", err)
	}
	_ = client.Close()
	h.activeConn.Delete(client)
}
func (p *pool) handleConn(conn net.Conn) {
	if p.closed.Load() {
		//如果已经关闭了，就不要处理这个链接了
		_ = conn.Close()
		return
	}
	client := connection.NewConn(conn, p.ConnWriteTimeout, p.ReadTimeout, p.KeepAliveTimeout)
	p.activeConn.Store(client, struct{}{})
	ch := parse.ParseStream(conn)
	for payload := range ch {
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// connection closed
				p.closeClient(client)
				logger.Error("connection closed: " + client.RemoteAddr().String())
				return
			}
			// protocol err
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				p.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		result := p.db.Exec(client, r.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}

}
