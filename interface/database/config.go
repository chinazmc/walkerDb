package databaseface

type StandaloneDatabaseConfig struct {
	DataDir string
}
type RaftDatabaseConfig struct {
	DataDir        string
	RaftDataDir    string
	RaftTCPAddress string // construct Raft Address
	TCPAddress     string
	NodeId         string
	JoinAddress    string // peer address to join
}
