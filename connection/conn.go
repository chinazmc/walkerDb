package connection

import (
	"net"
	"sync"
	"time"
)

type Connection struct {
	conn net.Conn
	//waiting until reply finished
	waitingReply Wait
	mu           sync.Mutex
}

func NewConn(conn net.Conn, writeTimeout, readTimeout, keepAliveTimeout time.Duration) *Connection {
	if readTimeout != 0 {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
	}
	if writeTimeout != 0 {
		conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	}
	if keepAliveTimeout != 0 {
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(keepAliveTimeout)
		}
	}
	return &Connection{
		conn: conn,
	}
}
func (c *Connection) GetConn() net.Conn {
	return c.conn
}
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}
func (c *Connection) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()
	_, err := c.conn.Write(b)
	return err
}

// Wait 与sync.WaitGroup类似，可以超时等待
type Wait struct {
	wg sync.WaitGroup
}

// Add 将增量（可能为负数）添加到WaitGroup计数器。
func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

// Done 将WaitGroup计数器减少一
func (w *Wait) Done() {
	w.wg.Done()
}

// Wait 直到WaitGroup计数器为零。
func (w *Wait) Wait() {
	w.wg.Wait()
}

// WaitWithTimeout 阻塞，直到WaitGroup计数器为零或超时
// returns 如果超时return ture
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan bool, 1)
	go func() {
		defer close(c)
		w.wg.Wait()
		c <- true
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
