package cluster

import (
	"fmt"
	"github.com/hashicorp/raft"
	"net/http"
	"runtime/debug"
	"strings"
	"walkerDb/database"
	"walkerDb/logger"
	"walkerDb/reply"
	"walkerDb/tcp"
)

type RaftDatabase struct {
	db   *database.DB
	raft *RaftNodeInfo
}
type ConsistencyLevel string

// Represents the available consistency levels.
const (
	Default    ConsistencyLevel = "0"
	Stale      ConsistencyLevel = "1"
	Consistent ConsistencyLevel = "2"
)

func (d *RaftDatabase) Close() {}
func (d *RaftDatabase) Set(key []byte, value []byte) error {

}
func (d *RaftDatabase) Get(key []byte) ([]byte, error) {

}

// Del deletes an item in the cache by key and returns true or false if a delete occurred.
func (d *RaftDatabase) Del(key []byte) (error, bool) {

}
func (d *RaftDatabase) SetEX(key []byte, value []byte, expireSeconds int) error {

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
	case "join":
		peerAddress := string(cmdLine[1])
		addPeerFuture := d.raft.raft.AddVoter(raft.ServerID(peerAddress), raft.ServerAddress(peerAddress), 0, 0)
		if err := addPeerFuture.Error(); err != nil {
			logger.Error(fmt.Sprintf("Error joining peer to raft, peeraddress:%s, err:%v, code:%d", peerAddress, err, http.StatusInternalServerError))
			return tcp.InternalErr
		}
		return tcp.OK
	case "consistency_level_get":
		key := cmdLine[1]
		lvl := level(string(cmdLine[2]))
		res, err := d.ConsistencyLevelGet(key, lvl)
		if err != nil {
			return tcp.NIL
		}
		return res
	default:
		return reply.MakeErrReply("error operator type").ToBytes()
	}
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (d *RaftDatabase) join(nodeID, httpAddr string, addr string) error {
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
	if err := d.SetMeta([]byte(nodeID), []byte(httpAddr)); err != nil {
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
				return tcp.NotLeaderErr, nil
			}
			redirect := s.FormRedirect(r, leader)
			http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
			return
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
	case "default":
		return Default
	case "stale":
		return Stale
	case "consistent":
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
