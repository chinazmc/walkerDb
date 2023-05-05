package main

import (
	"fmt"
	"net"
	"walkerDb"
	"walkerDb/logger"
)

func main() {
	listener, err := net.Listen("tcp", ":8972")
	if err != nil {
		panic(err)
	}
	walkerDb.StartEpollAndWorkerPool(listener)
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

		if err := walkerDb.Epoller.Add(conn); err != nil {
			logger.Info(fmt.Sprintf("failed to add connection %v", err))
			conn.Close()
		}
	}

	walkerDb.WorkerPool.Close()
}
