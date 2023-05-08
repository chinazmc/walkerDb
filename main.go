package main

import (
	_ "walkerDb/logger"
	"walkerDb/tcp"
)

func main() {
	op := NewOptions()
	server := tcp.NewServer(&tcp.Config{
		RaftTCPAddress: op.raftTCPAddress,
		DataDir:        op.dataDir,
		Bootstrap:      op.bootstrap,
		JoinAddress:    op.joinAddress,
	})
	server.Start()
}
