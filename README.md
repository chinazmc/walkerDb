walkerDb
======

**walkerDb** is a fast KV store engine based on the bitcask model.

It is compatible with the Redis protocol and supports consistent raft clustering.


## Running walkerDb
*Building walkerDb requires Go 1.9 or later.*

Starting and running a walkerDb cluster is easy.

Build walkerDb like so:
```bash
go install ./...
```

#### Step 1: Modify the /etc/hosts File

Add your servers’ hostnames and IP addresses to each cluster server’s /etc/hosts file (the hostnames below are representative).

```
<Meta_1_IP> raft-cluster-host-01
<Meta_2_IP> raft-cluster-host-02
<Meta_3_IP> raft-cluster-host-03
```

> Verification steps:
>
> Before proceeding with the installation, verify on each server that the other servers are resolvable. Here is an example set of shell commands using ping:
>
> ping -qc 1 raft-cluster-host-01
>
> ping -qc 1 raft-cluster-host-02
>
> ping -qc 1 raft-cluster-host-03

We highly recommend that each server be able to resolve the IP from the hostname alone as shown here. Resolve any connectivity issues before proceeding with the installation. A healthy cluster requires that every meta node can communicate with every other meta node.

### Step 2: Bring up a cluster or a standalone database
#### if you want to bring up a cluster
Run your first walkerDb node like so:
```bash
$GOPATH/bin/walkerDb -id node01 -raft raft-cluster-host-01:8001 -tcp raft-cluster-host-01:8011 -raftdir ./tmp1/raft -dir ./tmp1/bitcask-go -bootstrap true ~/.walkerDb
```

Let's bring up 2 more nodes, so we have a 3-node cluster. That way we can tolerate the failure of 1 node:
```bash
$GOPATH/bin/walkerDb -id node02 -raft 127.0.0.1:8002 -tcp 127.0.0.1:8022 -raftdir ./tmp2/raft -dir ./tmp2/bitcask-go -join 127.0.0.1:8011 -bootstrap true ~/.walkerDb

$GOPATH/bin/walkerDb -id node03 -raft 127.0.0.1:8003 -tcp 127.0.0.1:8033 -raftdir ./tmp3/raft -dir ./tmp3/bitcask-go -join 127.0.0.1:8011 -bootstrap true ~/.walkerDb
```
#### if you want to bring up a standalone database
```bash
$GOPATH/bin/walkerDb -tcp raft-cluster-host-01:8011 -dir ./tmp1/bitcask-go -bootstrap false ~/.walkerDb
```

## Reading and writing keys
You can now set a key and read its value back:
```go
func main() {
    conn, err := net.Dial("tcp", "0.0.0.0:8033")
    if err != nil {
    fmt.Println("err :", err)
        return
    }
    defer conn.Close()
    req := reply.MakeMultiBulkReply([][]byte{
        []byte("set"),
        []byte("hello"),
        []byte("world"),
    })
    _, err = conn.Write(req.ToBytes()) // 发送数据
    if err != nil {
        return
    }
    buf := [512]byte{}
    n, err := conn.Read(buf[:])
    if err != nil {
    fmt.Println("recv failed, err:", err)
        return
    }
    fmt.Println(string(buf[:n]))
}
```

You can now delete a key and its value:
```go
func main() {
    conn, err := net.Dial("tcp", "0.0.0.0:8033")
    if err != nil {
    fmt.Println("err :", err)
        return
    }
    defer conn.Close()
    req := reply.MakeMultiBulkReply([][]byte{
        []byte("del"),
        []byte("hello"),
    })
    _, err = conn.Write(req.ToBytes()) // 发送数据
    if err != nil {
        return
    }
    buf := [512]byte{}
    n, err := conn.Read(buf[:])
    if err != nil {
    fmt.Println("recv failed, err:", err)
        return
    }
    fmt.Println(string(buf[:n]))
}
```

### Three read consistency level
You can now read the key's value by different read consistency level:
```go
const (
    Default    = "0" //Execute the command directly on the leader node
    Stale      = "1" //Execute commands on arbitrary nodes with no guarantee of data consistency
    Consistent = "2" //On the leader node, the draft protocol should be used to determine whether the current leader is legal before executing the command, if it is legal, the command will be executed, otherwise it returns an error.
)
func main() {
    conn, err := net.Dial("tcp", "0.0.0.0:8033")
    if err != nil {
    fmt.Println("err :", err)
        return
    }
    defer conn.Close()
    req := reply.MakeMultiBulkReply([][]byte{
        []byte("levelget"),
        []byte("my"),
        []byte(Stale),
    })
    _, err = conn.Write(req.ToBytes()) // 发送数据
    if err != nil {
        return
    }
    buf := [512]byte{}
    n, err := conn.Read(buf[:])
    if err != nil {
    fmt.Println("recv failed, err:", err)
        return
    }
    fmt.Println(string(buf[:n]))
}
```
## benchmark
```bash
goos: windows
goarch: amd64
pkg: walkerDb/benchmark
cpu: Intel(R) Core(TM) i5-4460  CPU @ 3.20GHz
Benchmark_Put-4            55725             30891 ns/op            4625 B/op          9 allocs/op
Benchmark_Get-4          1924330               628.5 ns/op           135 B/op          4 allocs/op
Benchmark_Delete-4       1961263               613.6 ns/op           135 B/op          4 allocs/op
PASS
ok      walkerDb/benchmark      7.261s

goos: linux
goarch: amd64
pkg: walkerDb/benchmark
cpu: Intel(R) Xeon(R) CPU E3-1230 v5 @ 3.40GHz
Benchmark_Put-8      	   88315	     14172 ns/op	    4629 B/op	      10 allocs/op
Benchmark_Get-8      	 2244297	       527.9 ns/op	     135 B/op	       4 allocs/op
Benchmark_Delete-8   	 2291954	       519.4 ns/op	     135 B/op	       4 allocs/op
PASS
ok  	walkerDb/benchmark	5.984s

```

## 🔮 What will I do next?

- [ ] Implement batch write with transaction semantics.
- [ ] Optimize hintfile storage structure to support the memtable build faster (may use gob).
- [ ] Increased use of flatbuffers build options to support faster reading speed.
- [ ] Use mmap to read data file that on disk. [ however, the official mmap library is not optimized enough and needs to be further optimized ]
- [ ] Extend to build complex data structures with the same interface as Redis, such as List, Hash, Set, ZSet, Bitmap, etc.

