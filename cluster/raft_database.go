package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
	"walkerDb/database"
	databaseface "walkerDb/interface/database"
	"walkerDb/logger"
	"walkerDb/parse"
	"walkerDb/reply"
	"walkerDb/utils"
)

type RaftDatabase struct {
	db   *database.DB
	raft *RaftNodeInfo
	//enableWrite int32
	config *databaseface.RaftDatabaseConfig
}
type ConsistencyLevel string

var (
	// ErrOpenTimeout is returned when the Store does not apply its initial
	// logs within the specified time.
	ErrOpenTimeout = errors.New("timeout waiting for initial logs application")
	ErrNotLeader   = errors.New("current node is not the leader")
)

// Represents the available consistency levels.
const (
	Default            ConsistencyLevel = "0"
	Stale              ConsistencyLevel = "1"
	Consistent         ConsistencyLevel = "2"
	ENABLE_WRITE_TRUE                   = int32(1)
	ENABLE_WRITE_FALSE                  = int32(0)
	leaderWaitDelay                     = 100 * time.Millisecond
	appliedWaitDelay                    = 100 * time.Millisecond
	raftTimeout                         = 10 * time.Second
)

func NewRaftDatabase(config *databaseface.RaftDatabaseConfig) *RaftDatabase {
	databaseOpts := database.DefaultOptions
	if config.DataDir != "" {
		databaseOpts.DirPath = config.DataDir
	}
	db, err := database.Open(databaseOpts)
	if err != nil {
		panic(err)
	}
	raftNode, err := newRaftNode(config, db)
	if err != nil {
		logger.Error(fmt.Sprintf("new raft node failed:%v", err))
	}
	raftDatabase := &RaftDatabase{db: db, raft: raftNode, config: config}
	return raftDatabase
}
func (d *RaftDatabase) Start() {
	time.Sleep(100 * time.Millisecond)
	config := d.config
	if config.JoinAddress != "" {
		err := joinRaftCluster(config)
		if err != nil {
			logger.Error(fmt.Sprintf("join raft cluster failed:%v", err))
		}
	}
	// Wait until the store is in full consensus.
	openTimeout := 120 * time.Second
	d.WaitForLeader(openTimeout)
	d.WaitForApplied(openTimeout)
	// This may be a standalone server. In that case set its own metadata.
	if err := d.SetMeta([]byte(config.NodeId), []byte(config.TCPAddress)); err != nil && err != ErrNotLeader {
		// Non-leader errors are OK, since metadata will then be set through
		// consensus as a result of a join. All other errors indicate a problem.
		log.Fatalf("failed to SetMeta at %s: %s", config.NodeId, err.Error())
	}
	go func() {
		// monitor leadership
		for {
			select {
			case leader := <-d.raft.leaderNotifyCh:
				if leader {
					logger.Info("become leader, enable write api")
					//d.setWriteFlag(true)
				} else {
					logger.Info("become follower, close write api")
					//d.setWriteFlag(false)
				}
			}
		}
	}()
}
func (d *RaftDatabase) redirect(req []byte) ([]byte, error) {
	//跳转到leader取处理
	leader := d.LeaderAPIAddr()
	if leader == "" {
		//如果还没有leader，那么就需要等待到有leader
		logger.Error("current stage not leader,waiting!!!")
		d.WaitForLeader(120 * time.Second)
		leader = d.LeaderAPIAddr()
		if leader == "" {
			return nil, ErrNotLeader
		}
	}
	res, err := utils.SendTcpReq(leader, req)
	if err != nil {
		return res, err
	}
	return res, nil
}
func (d *RaftDatabase) Close() { d.db.Close() }
func (d *RaftDatabase) Set(key []byte, value []byte) error {
	if d.raft.raft.State() != raft.Leader {
		req := reply.MakeMultiBulkReply([][]byte{
			[]byte("set"),
			key,
			value,
		})

		_, err := d.redirect(req.ToBytes())
		return err
	}
	c := &logEntryData{
		Op:    "set",
		Key:   key,
		Value: value,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f := d.raft.raft.Apply(b, raftTimeout)
	return f.Error()
}
func (d *RaftDatabase) Get(key []byte) ([]byte, error) {
	return d.ConsistencyLevelGet(key, Stale)
}

// Del deletes an item in the cache by key and returns true or false if a delete occurred.
func (d *RaftDatabase) Del(key []byte) (error, bool) {
	if d.raft.raft.State() != raft.Leader {
		req := reply.MakeMultiBulkReply([][]byte{
			[]byte("del"),
			key,
		})

		_, err := d.redirect(req.ToBytes())
		if err != nil {
			return err, false
		}
		return nil, true
	}

	c := &logEntryData{
		Op:  "del",
		Key: key,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err, false
	}

	f := d.raft.raft.Apply(b, raftTimeout)
	if f.Error() != nil {
		return f.Error(), false
	}
	return nil, true
}
func (d *RaftDatabase) SetEX(key []byte, value []byte, expireSeconds int) error {

	return d.Set(key, value)
}
func (d *RaftDatabase) Exec(cmdLine [][]byte) (result []byte) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = []byte("-Err unknown\r\n")
		}
	}()

	cmdName := strings.ToLower(string(cmdLine[0]))
	switch cmdName {
	case string(parse.JOIN):
		nodeId := string(cmdLine[1])
		peerAddress := string(cmdLine[2])
		peertcpAddress := string(cmdLine[3])
		err := d.join(nodeId, peerAddress, peertcpAddress)
		if err != nil {
			logger.Error(fmt.Sprintf("Error joining peer to raft, peeraddress:%s, err:%v, code:%d", peerAddress, err, http.StatusInternalServerError))
			return parse.NotLeaderErr
		}
		return parse.OK
	case string(parse.LEVELGET):
		key := cmdLine[1]
		lvl := level(string(cmdLine[2]))
		res, err := d.ConsistencyLevelGet(key, lvl)
		if err != nil {
			return parse.NIL
		}
		return res
	default:
		return reply.MakeErrReply("error operator type").ToBytes()
	}
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (d *RaftDatabase) join(nodeID, addr, tcpAddr string) error {
	if d.raft.raft.State() != raft.Leader {
		req := reply.MakeMultiBulkReply([][]byte{
			[]byte("join"),
			[]byte(nodeID),
			[]byte(addr),
			[]byte(tcpAddr),
		})

		_, err := d.redirect(req.ToBytes())
		if err != nil {
			return err
		}
		return nil
	}
	logger.Info(fmt.Sprintf("received join request for remote node %s at %s", nodeID, addr))

	configFuture := d.raft.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		logger.Error(fmt.Sprintf("failed to get raft configuration: %v", err))
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				logger.Info("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := d.raft.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := d.raft.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	// Set meta info
	if err := d.SetMeta([]byte(nodeID), []byte(tcpAddr)); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("node %s at %s joined successfully", nodeID, addr))
	return nil
}
func (d *RaftDatabase) SetMeta(key, value []byte) error {
	return d.Set(key, value)
}

func (d *RaftDatabase) GetMeta(key []byte) (string, error) {
	res, err := d.ConsistencyLevelGet(key, Stale)
	return string(res), err
}

func (d *RaftDatabase) DeleteMeta(key []byte) error {
	err, _ := d.Del(key)
	return err
}
func (d *RaftDatabase) ConsistencyLevelGet(key []byte, level ConsistencyLevel) ([]byte, error) {
	if level != Stale {
		if d.raft.raft.State() != raft.Leader {
			//跳转到leader取处理
			leader := d.LeaderAPIAddr()
			if leader == "" {
				return parse.NotLeaderErr, nil
			}
			req := reply.MakeMultiBulkReply([][]byte{
				[]byte("levelget"),
				key,
				[]byte(level),
			})
			res, err := utils.SendTcpReq(leader, req.ToBytes())
			if err != nil {
				return nil, err
			}
			return res, nil
		}

	}
	//在不是stale读的情况下，只有leader才能进到这里，那么default的话就不进行一致性读，直接读leader的本地数据就行
	if level == Consistent {
		if err := d.consistentRead(); err != nil {
			return reply.MakeErrReply(err.Error()).ToBytes(), nil
		}
	}

	return d.db.Get(key)
}
func (s *RaftDatabase) LeaderAPIAddr() string {
	id, err := s.LeaderID()
	if err != nil {
		return ""
	}

	addr, err := s.GetMeta([]byte(id))
	if err != nil {
		return ""
	}

	return addr
}

// LeaderID returns the node ID of the Raft leader. Returns a
// blank string if there is no leader, or an error.
func (s *RaftDatabase) LeaderID() (string, error) {
	addr := s.LeaderAddr()
	configFuture := s.raft.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		logger.Error(fmt.Sprintf("failed to get raft configuration: %v", err))
		return "", err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.Address == raft.ServerAddress(addr) {
			return string(srv.ID), nil
		}
	}
	return "", nil
}
func (s *RaftDatabase) LeaderAddr() string {
	return string(s.raft.raft.Leader())
}
func level(lvl string) ConsistencyLevel {
	switch strings.ToLower(lvl) {
	case "0":
		return Default
	case "1":
		return Stale
	case "2":
		return Consistent
	default:
		return Default
	}
}

// consistentRead is used to ensure we do not perform a stale
// read. This is done by verifying leadership before the read.
func (s *RaftDatabase) consistentRead() error {
	future := s.raft.raft.VerifyLeader()
	if err := future.Error(); err != nil {
		return err //fail fast if leader verification fails
	}

	return nil
}

// WaitForAppliedIndex blocks until a given log index has been applied,
// or the timeout expires.
func (s *RaftDatabase) WaitForAppliedIndex(idx uint64, timeout time.Duration) error {
	tck := time.NewTicker(appliedWaitDelay)
	defer tck.Stop()
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	for {
		select {
		case <-tck.C:
			if s.raft.raft.AppliedIndex() >= idx {
				return nil
			}
		case <-tmr.C:
			return fmt.Errorf("timeout expired")
		}
	}
}

// WaitForApplied waits for all Raft log entries to to be applied to the
// underlying database.
func (s *RaftDatabase) WaitForApplied(timeout time.Duration) error {
	if timeout == 0 {
		return nil
	}
	logger.Info("waiting for up to %s for application of initial logs", timeout)
	if err := s.WaitForAppliedIndex(s.raft.raft.LastIndex(), timeout); err != nil {
		return ErrOpenTimeout
	}
	return nil
}

// WaitForLeader blocks until a leader is detected, or the timeout expires.
func (s *RaftDatabase) WaitForLeader(timeout time.Duration) (string, error) {
	tck := time.NewTicker(leaderWaitDelay)
	defer tck.Stop()
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	for {
		select {
		case <-tck.C:
			l := s.LeaderAddr()
			if l != "" {
				return l, nil
			}
		case <-tmr.C:
			return "", fmt.Errorf("timeout expired")
		}
	}
}

//func (s *RaftDatabase) checkWritePermission() bool {
//	return atomic.LoadInt32(&s.enableWrite) == ENABLE_WRITE_TRUE
//}
//
//func (s *RaftDatabase) setWriteFlag(flag bool) {
//	if flag {
//		atomic.StoreInt32(&s.enableWrite, ENABLE_WRITE_TRUE)
//	} else {
//		atomic.StoreInt32(&s.enableWrite, ENABLE_WRITE_FALSE)
//	}
//}
