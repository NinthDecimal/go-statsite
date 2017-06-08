package statsite

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// Network represets a set of servers connected together
type Network interface {
	ResolveTCPAddr(connType string, address string) error
	DialTimeout(connType string, address string, timeout time.Duration) (net.Conn, error)
}

type realNetwork struct{}

// NewRealNetwork returns a real network pointer
func NewRealNetwork() Network {
	return &realNetwork{}
}

// ResolveTCPAddr resolves the tcp address host:port in the address string
func (t realNetwork) ResolveTCPAddr(connType string, address string) error {
	_, err := net.ResolveTCPAddr(connType, address)
	return err
}

// DialTimeout connects to the tcp address
func (t realNetwork) DialTimeout(connType string, address string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout(connType, address, timeout)
	return conn, err
}

type mockNetwork struct {
	mockServers map[string]mockServer
}

// NewMockNetwork returns a new mock network that is able to connect to any
// mockServer in the mockServer map
func NewMockNetwork(serverMap map[string]mockServer) Network {
	return &mockNetwork{
		mockServers: serverMap,
	}
}

// ResolveTCPAddr mocks the resolve coll and returns an error if the address
// is the string "invalid"
func (t mockNetwork) ResolveTCPAddr(connType string, address string) error {
	if address == "invalid" {
		return errors.New("Invalid address")
	}
	return nil
}

// DialTimeout mocks the connect call and returns an error if the address does
// not match a mockServer in the network map
func (t mockNetwork) DialTimeout(connType string, address string, timeout time.Duration) (net.Conn, error) {
	s := t.mockServers[address]
	if s == nil {
		return nil, fmt.Errorf("Server not found on mockNetwork %s", address)
	}
	connection := &mockConnection{server: s}
	return connection, nil
}

// Client represents a means of emitting Messages to a server
type Client interface {
	Emit(msg Message) error
	Connect() error
	Close()
}

// client is an implementation of the Client interface for connecting and
// Emitting metrics
type client struct {
	Conn    net.Conn
	addr    string
	network Network
}

// NewNetworkClient takes an address string in the form "host:port" and
// a Network and returns a Client
func NewNetworkClient(addr string, network Network) Client {
	return &client{
		Conn:    nil,
		addr:    addr,
		network: network,
	}
}

// NewClient takes an address string in the form "host:port" and returns a
// Client on a realNetwork
func NewClient(addr string) Client {
	network := &realNetwork{}
	return NewNetworkClient(addr, network)
}

// Connect instructs a Client to make a connection to the server at the address
// specified in the client
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

// Close cleans up any connections
func (t *client) Close() {
	t.Conn = nil
}

// Emit sends a message to the address defined in the client
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

// Read returns the number of messages sent to the mockConnection
func (t *mockConnection) Read(b []byte) (n int, err error) {
	return int(t.server.Count()), nil
}

// Write sends data to a mockConnection
func (t *mockConnection) Write(b []byte) (n int, err error) {
	s := string(b)
	err = t.server.Write(s)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

// Close is a no-op for mockConnection
func (t *mockConnection) Close() error {
	return nil
}

// LocalAddr returns a mockAddress needed to implement a net.Conn
func (t *mockConnection) LocalAddr() net.Addr {
	return &mockAddr{
		network: "tcp",
		address: "localhost",
	}
}

// RemoteAddr returns a mockAddress needed to implement a net.Conn
func (t *mockConnection) RemoteAddr() net.Addr {
	return &mockAddr{
		network: "tcp",
		address: "mockremote",
	}
}

// SetDeadline is a no-op for mockConnection
func (t *mockConnection) SetDeadline(time time.Time) error {
	return nil
}

// SetReadDeadline is a no-op for mockConnection
func (t *mockConnection) SetReadDeadline(time time.Time) error {
	return nil
}

// SetWriteDeadline is a no-op for mockConnection
func (t *mockConnection) SetWriteDeadline(time time.Time) error {
	return nil
}

type mockAddr struct {
	network string
	address string
}

// Network returns the network connection type
func (t mockAddr) Network() string {
	return t.network
}

// String prints the address
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

// Write takes a string and writes to the local cache for the mockStatsite
func (t *mockStatsite) Write(s string) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if s == "bad:key|kv\n" {
		return errors.New("Error writing to statstie")
	}
	t.receivedMessages = append(t.receivedMessages, s)
	return nil
}

// Read returns all the messages sent to the mockStatsite
func (t *mockStatsite) Read() []string {
	return t.receivedMessages
}

// Last returns the last message sent to the mockStatsite
func (t *mockStatsite) Last() string {
	t.lock.Lock()
	defer t.lock.Unlock()
	if len(t.receivedMessages) == 0 {
		return ""
	}
	return t.receivedMessages[len(t.receivedMessages)-1]
}

// Count returns the total number of messages sent to the mockStatsite
func (t *mockStatsite) Count() int {
	t.lock.Lock()
	defer t.lock.Unlock()

	return len(t.receivedMessages)
}
