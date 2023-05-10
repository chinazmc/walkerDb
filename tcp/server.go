package tcp

import (
	"fmt"
	"net"
	"time"
	"walkerDb/cluster"
	"walkerDb/database"
	databaseface "walkerDb/interface/database"
	"walkerDb/logger"
)

// Config stores tcp server properties
type Config struct {
	RaftTCPAddress string
	DataDir        string
	Bootstrap      bool
	JoinAddress    string
	TcpAddress     string
	RaftDataDir    string
	NodeId         string
}
type Server struct {
	cache  databaseface.Database
	option *Config
	//activeConn sync.Map // *client -> placeholder
	workPool         *pool
	ConnWriteTimeout time.Duration
	ReadTimeout      time.Duration
	KeepAliveTimeout time.Duration
	Deadline         time.Duration
}

func NewServer(config *Config) (server *Server) {
	server = &Server{
		option:   config,
		Deadline: 60 * time.Second,
	}
	if !server.option.Bootstrap {
		server.cache = database.NewStandaloneDatabase(databaseface.StandaloneDatabaseConfig{
			DataDir: config.DataDir,
		})
	} else {
		//cluster
		server.cache = cluster.NewRaftDatabase(&databaseface.RaftDatabaseConfig{
			DataDir:        config.DataDir,
			RaftDataDir:    config.RaftDataDir,
			RaftTCPAddress: config.RaftTCPAddress,
			TCPAddress:     config.TcpAddress,
			NodeId:         config.NodeId,
			JoinAddress:    config.JoinAddress,
		})
	}
	return server
}
func (s *Server) Start() {
	listener, err := net.Listen("tcp", s.option.TcpAddress)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	logger.Info("Listening on port 8972")
	s.StartWorkerPool()
	s.StartEpoll(listener)
	go s.cache.Start()
	for {
		conn, e := listener.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Timeout() {
				logger.Info(fmt.Sprintf("accept temp err: %v", ne))
				continue
			}

			logger.Info(fmt.Sprintf("accept err: %v", e))
			return
		}

		if err := Epoller.Add(conn); err != nil {
			logger.Info(fmt.Sprintf("failed to add connection %v", err))
			conn.Close()
		}
	}

	WorkerPool.Close()
}
func (s *Server) StartWorkerPool() {
	WorkerPool = newPool(*c, 1000000)
	WorkerPool.start()
	s.workPool = WorkerPool
}
func (s *Server) StartEpoll(listener net.Listener) {
	var err error
	Epoller, err = MkEpoll(listener)
	if err != nil {
		panic(err)
	}

	go func() {
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
				client := NewConn(conn, s)
				//s.activeConn.Store(client, struct{}{})
				go client.readLoop()
			}
		}
	}()
}
