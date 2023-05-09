package main

import (
	"walkerDb/cluster"
	databaseface "walkerDb/interface/database"
)

type cache struct {
	//opts *options
	db   *databaseface.Database
	raft *cluster.RaftNodeInfo
}

type CacheContext struct {
	c *cache
}

//
//func main() {
//	st := &cache{
//		opts: NewOptions(),
//	}
//	if !st.opts.bootstrap {
//		//单机启动
//		listener, err := net.Listen("tcp", ":8972")
//		if err != nil {
//			panic(err)
//		}
//		walkerDb.StartEpollAndWorkerPool(listener)
//		for {
//			conn, e := listener.Accept()
//			if e != nil {
//				if ne, ok := e.(net.Error); ok && ne.Timeout() {
//					logger.Info(fmt.Sprintf("accept temp err: %v", ne))
//					continue
//				}
//
//				logger.Info(fmt.Sprintf("accept err: %v", e))
//				return
//			}
//
//			if err := walkerDb.Epoller.Add(conn); err != nil {
//				logger.Info(fmt.Sprintf("failed to add connection %v", err))
//				conn.Close()
//			}
//		}
//
//		walkerDb.WorkerPool.Close()
//		return
//	}
//
//	//raft模式启动
//	st.db
//}
