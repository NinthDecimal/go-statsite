package statsite

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type Network interface {
	ResolveTCPAddr(connType string, address string) error
	DialTimeout(connType string, address string, timeout time.Duration) (net.Conn, error)
}

type realNetwork struct{}

func NewRealNetwork() Network {
	return &realNetwork{}
}

func (t realNetwork) ResolveTCPAddr(connType string, address string) error {
	_, err := net.ResolveTCPAddr(connType, address)
	return err
}

func (t realNetwork) DialTimeout(connType string, address string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout(connType, address, timeout)
	return conn, err
}

type mockNetwork struct{}

func NewMockNetwork() Network {
	return &mockNetwork{}
}

func (t mockNetwork) ResolveTCPAddr(connType string, address string) error {
	if address == "invalid" {
		return errors.New("Invalid address")
	}
	return nil
}

func (t mockNetwork) DialTimeout(connType string, address string, timeout time.Duration) (net.Conn, error) {
	if address == "badconnection" {
		return nil, errors.New("Unable to connect")
	}
	mockStatsite := &mockStatsite{
		receivedMessages: 0,
	}
	connection := &mockConnection{statsite: mockStatsite}
	return connection, nil
}

type Client interface {
	Emit(msg Message) error
	Connection() net.Conn
}

type client struct {
	Conn net.Conn
	addr string
}

func NewNetworkClient(addr string, network Network) (Client, error) {
	err := network.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Error resolving statsite: %v", err)
	}

	var conn net.Conn
	conn, err = network.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to statsite: %v", err)
	}

	return &client{
		Conn: conn,
		addr: addr,
	}, nil
}

func NewClient(addr string) (Client, error) {
	network := &realNetwork{}
	return NewNetworkClient(addr, network)

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

func (t client) Connection() net.Conn {
	return t.Conn
}

type mockConnection struct {
	statsite *mockStatsite
}

func (t mockConnection) Read(b []byte) (n int, err error) {
	return t.statsite.Read(), nil
}

func (t mockConnection) Write(b []byte) (n int, err error) {
	s := string(b)
	if s == "bad:key|kv\n" {
		return 0, errors.New("Error writing to statstie")
	}
	t.statsite.Write(s)
	return 1, nil
}

func (t mockConnection) Close() error {
	return nil
}

func (t mockConnection) LocalAddr() net.Addr {
	return &mockAddr{
		network: "tcp",
		address: "localhost",
	}
}

func (t mockConnection) RemoteAddr() net.Addr {
	return &mockAddr{
		network: "tcp",
		address: "statsite",
	}
}

func (t mockConnection) SetDeadline(time time.Time) error {
	return nil
}

func (t mockConnection) SetReadDeadline(time time.Time) error {
	return nil
}
func (t mockConnection) SetWriteDeadline(time time.Time) error {
	return nil
}

type mockAddr struct {
	network string
	address string
}

func (t mockAddr) Network() string {
	return t.network
}

func (t mockAddr) String() string {
	return t.address
}

type mockStatsite struct {
	receivedMessages int
}

func (t *mockStatsite) Write(s string) {
	t.receivedMessages++
}

func (t *mockStatsite) Read() int {
	return t.receivedMessages
}
