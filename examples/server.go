package main

import (
	"log"
	"net"
	"walkerDb"
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
				log.Printf("accept temp err: %v", ne)
				continue
			}

			log.Printf("accept err: %v", e)
			return
		}

		if err := walkerDb.Epoller.Add(conn); err != nil {
			log.Printf("failed to add connection %v", err)
			conn.Close()
		}
	}

	walkerDb.WorkerPool.Close()
}
