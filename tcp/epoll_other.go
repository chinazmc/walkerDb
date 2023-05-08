//go:build !linux
// +build !linux

package tcp

import (
	"net"
	"sync"
)

type epoll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
	listener    net.Listener
	ch          chan net.Conn
}

func MkEpoll(listener net.Listener) (*epoll, error) {

	return &epoll{
		listener: listener,
		ch:       make(chan net.Conn),
	}, nil
}

func (e *epoll) Add(conn net.Conn) error {
	e.ch <- conn
	return nil
}

func (e *epoll) Remove(conn net.Conn) error {
	return nil
}

func (e *epoll) Wait() ([]net.Conn, error) {
	conn := <-e.ch
	return []net.Conn{conn}, nil
}
