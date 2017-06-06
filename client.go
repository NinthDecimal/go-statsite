package statsite

import (
	"errors"
	"fmt"
	"net"
	"sync"
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

type mockNetwork struct {
	mockServers map[string]mockServer
}

func NewMockNetwork(serverMap map[string]mockServer) Network {
	return &mockNetwork{
		mockServers: serverMap,
	}
}

func (t mockNetwork) ResolveTCPAddr(connType string, address string) error {
	if address == "invalid" {
		return errors.New("Invalid address")
	}
	return nil
}

func (t mockNetwork) DialTimeout(connType string, address string, timeout time.Duration) (net.Conn, error) {
	s := t.mockServers[address]
	if s == nil {
		return nil, fmt.Errorf("Server not found on mockNetwork %s", address)
	}
	connection := &mockConnection{server: s}
	return connection, nil
}

type Client interface {
	Emit(msg Message) error
	Connect() error
	Close()
}

type client struct {
	Conn    net.Conn
	addr    string
	network Network
}

func NewNetworkClient(addr string, network Network) Client {
	return &client{
		Conn:    nil,
		addr:    addr,
		network: network,
	}
}

func NewClient(addr string) Client {
	network := &realNetwork{}
	return NewNetworkClient(addr, network)
}

func (t *client) Connect() error {
	err := t.network.ResolveTCPAddr("tcp", t.addr)
	if err != nil {
		return fmt.Errorf("Error resolving statsite: %v", err)
	}

	conn, err := t.network.DialTimeout("tcp", t.addr, 1*time.Second)
	if err != nil {
		return fmt.Errorf("Error connecting to statsite: %v", err)
	}
	t.Conn = conn
	return nil

}

func (t *client) Close() {
	// t.Conn = nil
}

func (t *client) Emit(msg Message) error {
	if t.Conn == nil {
		err := t.Connect()
		if err != nil {
			return err
		}
	}
	return t.emitter(msg.String())
}

func (t *client) emitter(msg string) error {
	_, err := t.Conn.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

type mockConnection struct {
	server mockServer
}

func (t *mockConnection) Read(b []byte) (n int, err error) {
	return int(t.server.Count()), nil
}

func (t *mockConnection) Write(b []byte) (n int, err error) {
	s := string(b)
	err = t.server.Write(s)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (t *mockConnection) Close() error {
	return nil
}

func (t *mockConnection) LocalAddr() net.Addr {
	return &mockAddr{
		network: "tcp",
		address: "localhost",
	}
}

func (t *mockConnection) RemoteAddr() net.Addr {
	return &mockAddr{
		network: "tcp",
		address: "statsite",
	}
}

func (t *mockConnection) SetDeadline(time time.Time) error {
	return nil
}

func (t *mockConnection) SetReadDeadline(time time.Time) error {
	return nil
}
func (t *mockConnection) SetWriteDeadline(time time.Time) error {
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

type mockServer interface {
	Write(string) error
	Read() []string
	Last() string
	Count() int
}

type mockStatsite struct {
	receivedMessages []string
	lock             sync.Mutex
}

func (t *mockStatsite) Write(s string) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if s == "bad:key|kv\n" {
		return errors.New("Error writing to statstie")
	}
	t.receivedMessages = append(t.receivedMessages, s)
	return nil
}

func (t *mockStatsite) Read() []string {
	return t.receivedMessages
}

func (t *mockStatsite) Last() string {
	t.lock.Lock()
	defer t.lock.Unlock()
	if len(t.receivedMessages) == 0 {
		return ""
	}
	return t.receivedMessages[len(t.receivedMessages)-1]
}

func (t *mockStatsite) Count() int {
	t.lock.Lock()
	defer t.lock.Unlock()

	return len(t.receivedMessages)
}
