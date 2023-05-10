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
		TcpAddress:     op.tcpAddress,
		RaftDataDir:    op.raftDataDir,
		NodeId:         op.nodeId,
	})
	server.Start()
}
