package cluster

//
//import "hash/crc32"
//
////https://www.cnblogs.com/Finley/p/14038398.html
//import (
//	"sort"
//	"strconv"
//)
//
//func main() {
//
//}
//
//type HashFunc func(data []byte) uint32
//type Map struct {
//	hashFunc HashFunc
//	replicas int
//	keys     []int // sorted
//	hashMap  map[int]string
//}
//
//func New(replicas int, fn HashFunc) *Map {
//	m := &Map{
//		replicas: replicas, // 每个物理节点会产生 replicas 个虚拟节点
//		hashFunc: fn,
//		hashMap:  make(map[int]string), // 虚拟节点 hash 值到物理节点地址的映射
//	}
//	if m.hashFunc == nil {
//		m.hashFunc = crc32.ChecksumIEEE
//	}
//	return m
//}
//
//func (m *Map) IsEmpty() bool {
//	return len(m.keys) == 0
//}
//func (m *Map) Add(keys ...string) {
//	for _, key := range keys {
//		if key == "" {
//			continue
//		}
//		for i := 0; i < m.replicas; i++ {
//			// 使用 i + key 作为一个虚拟节点，计算虚拟节点的 hash 值
//			hash := int(m.hashFunc([]byte(strconv.Itoa(i) + key)))
//			// 将虚拟节点添加到环上
//			m.keys = append(m.keys, hash)
//			// 注册虚拟节点到物理节点的映射
//			m.hashMap[hash] = key
//		}
//	}
//	sort.Ints(m.keys)
//}
//func (m *Map) Get(key string) string {
//	if m.IsEmpty() {
//		return ""
//	}
//
//	// 支持根据 key 的 hashtag 来确定分布
//	partitionKey := getPartitionKey(key)
//	hash := int(m.hashFunc([]byte(partitionKey)))
//
//	// sort.Search 会使用二分查找法搜索 keys 中满足 m.keys[i] >= hash 的最小 i 值
//	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })
//
//	// 若 key 的 hash 值大于最后一个虚拟节点的 hash 值，则 sort.Search 找不到目标
//	// 这种情况下选择第一个虚拟节点
//	if idx == len(m.keys) {
//		idx = 0
//	}
//
//	// 将虚拟节点映射为实际地址
//	return m.hashMap[m.keys[idx]]
//}
//
//// 集群模式下，除了 MSet、DEL 等特殊指令外，其它指令会交由 defaultFunc 处理
//func defaultFunc(cluster *Cluster, c redis.Connection, args [][]byte) redis.Reply {
//	key := string(args[1])
//	peer := cluster.peerPicker.Get(key) // 通过一致性 hash 找到节点
//	return cluster.Relay(peer, c, args)
//}
//
//func (cluster *Cluster) Relay(peer string, c redis.Connection, args [][]byte) redis.Reply {
//	if peer == cluster.self { // 若数据在本地则直接调用数据库引擎
//		// to self db
//		return cluster.db.Exec(c, args)
//	} else {
//		// 从连接池取一个与目标节点的连接
//		// 连接池使用 github.com/jolestar/go-commons-pool/v2 实现
//		peerClient, err := cluster.getPeerClient(peer)
//		if err != nil {
//			return reply.MakeErrReply(err.Error())
//		}
//		defer func() {
//			_ = cluster.returnPeerClient(peer, peerClient) // 处理完成后将连接放回连接池
//		}()
//		// 将指令发送到目标节点
//		return peerClient.Send(args)
//	}
//}
//
//func (cluster *Cluster) getPeerClient(peer string) (*client.Client, error) {
//	connectionFactory, ok := cluster.peerConnection[peer]
//	if !ok {
//		return nil, errors.New("connection factory not found")
//	}
//	raw, err := connectionFactory.BorrowObject(context.Background())
//	if err != nil {
//		return nil, err
//	}
//	conn, ok := raw.(*client.Client)
//	if !ok {
//		return nil, errors.New("connection factory make wrong type")
//	}
//	return conn, nil
//}
//
//func (cluster *Cluster) returnPeerClient(peer string, peerClient *client.Client) error {
//	connectionFactory, ok := cluster.peerConnection[peer]
//	if !ok {
//		return errors.New("connection factory not found")
//	}
//	return connectionFactory.ReturnObject(context.Background(), peerClient)
//}
