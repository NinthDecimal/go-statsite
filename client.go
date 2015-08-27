package statsite

import (
	"fmt"
	"net"
	"time"
)

type Client interface {
	Emit(msg Message) error
}

type client struct {
	Conn net.Conn
	addr string
}

func (t client) Emit(msg Message) error {
	return t.emitter(msg.String())
}

func (t client) emitter(msg string) error {
	_, err := t.Conn.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func NewClient(addr string) (Client, error) {

	_, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Error resolving statsite: %v", err)
	}

	var conn net.Conn
	conn, err = net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to statsite: %v", err)
	}

	return &client{
		Conn: conn,
		addr: addr,
	}, nil
}
