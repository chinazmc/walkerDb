package cluster

import (
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"net"
	"os"
	"path/filepath"
	"time"
	"walkerDb/database"
	databaseface "walkerDb/interface/database"
	"walkerDb/parse"
	"walkerDb/reply"
	"walkerDb/utils"
)

type RaftNodeInfo struct {
	raft           *raft.Raft
	fsm            *FSM
	leaderNotifyCh chan bool
	nodeId         string
	serverAddress  string
}

func newRaftTransport(config *databaseface.RaftDatabaseConfig) (*raft.NetworkTransport, error) {
	address, err := net.ResolveTCPAddr("tcp", config.RaftTCPAddress)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}
	return transport, nil
}
func newRaftNode(config *databaseface.RaftDatabaseConfig, db *database.DB) (*RaftNodeInfo, error) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(config.NodeId)
	//raftConfig.Logger = log.New(os.Stderr, "raft: ", log.Ldate|log.Ltime)
	raftConfig.SnapshotInterval = 20 * time.Second
	raftConfig.SnapshotThreshold = 2 //每commit多少log entry后生成一次快照
	leaderNotifyCh := make(chan bool, 1)
	raftConfig.NotifyCh = leaderNotifyCh

	transport, err := newRaftTransport(config)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(config.RaftDataDir, os.ModePerm); err != nil {
		return nil, err
	}

	fsm := &FSM{
		db: db,
	}
	snapshotStore, err := raft.NewFileSnapshotStore(config.RaftDataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(config.RaftDataDir, "raft-log.bolt"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(config.RaftDataDir, "raft-stable.bolt"))
	if err != nil {
		return nil, err
	}

	raftNode, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raftConfig.LocalID,
				Address: transport.LocalAddr(),
			},
		},
	}
	raftNode.BootstrapCluster(configuration)
	return &RaftNodeInfo{raft: raftNode, fsm: fsm, leaderNotifyCh: leaderNotifyCh, nodeId: config.NodeId,
		serverAddress: config.RaftTCPAddress}, nil
}

// joinRaftCluster joins a node to raft cluster
func joinRaftCluster(opts *databaseface.RaftDatabaseConfig) error {
	req := reply.MakeMultiBulkReply([][]byte{
		[]byte("join"),
		[]byte(opts.NodeId),
		[]byte(opts.RaftTCPAddress),
		[]byte(opts.TCPAddress),
	})
	res, err := utils.SendTcpReq(opts.JoinAddress, req.ToBytes())
	if err != nil {
		return err
	}
	if string(res) != string(parse.OK) {
		return errors.New(fmt.Sprintf("Error joining cluster: %s", res))
	}
	return nil
}
