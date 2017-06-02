package statsite

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type ClientSuite struct {
	mockNetwork Network
}

var _ = Suite(&ClientSuite{})

func (s *ClientSuite) SetUpTest(c *C) {
	s.mockNetwork = NewMockNetwork()
}

func (s *ClientSuite) TestNewClient(c *C) {
	client, err := NewNetworkClient("localhost:80", s.mockNetwork)
	c.Assert(err, IsNil)
	c.Assert(client.Connection(), NotNil)
}

func (s *ClientSuite) TestNewClientInvalid(c *C) {
	_, err := NewNetworkClient("invalid", s.mockNetwork)
	c.Assert(err, ErrorMatches, "Error resolving statsite: Invalid address")
}

func (s *ClientSuite) TestNewClientConnect(c *C) {
	_, err := NewNetworkClient("badconnection", s.mockNetwork)
	c.Assert(err, ErrorMatches, "Error connecting to statsite: Unable to connect")
}

func (s *ClientSuite) TestEmit(c *C) {
	client, err := NewNetworkClient("mockStatsite", s.mockNetwork)
	c.Assert(err, IsNil)
	conn := client.Connection()
	msg := NewKeyValue("key", "value")
	// test 0 sent
	b := []byte{}
	sent, err := conn.Read(b)
	c.Assert(err, IsNil)
	c.Assert(sent, Equals, 0)
	err = client.Emit(msg)
	c.Assert(err, IsNil)
	// Expect 1 stat added to statsite
	sent, err = conn.Read(b)
	c.Assert(err, IsNil)
	c.Assert(sent, Equals, 1)
}

func (s *ClientSuite) TestEmitFailure(c *C) {
	client, err := NewNetworkClient("mockStatsite", s.mockNetwork)
	c.Assert(err, IsNil)
	conn := client.Connection()
	msg := NewKeyValue("bad", "key")
	// test 0 sent
	b := []byte{}
	sent, err := conn.Read(b)
	c.Assert(err, IsNil)
	c.Assert(sent, Equals, 0)
	// Expect error
	err = client.Emit(msg)
	c.Assert(err, NotNil)
	// Expect no stats added to statsite
	sent, err = conn.Read(b)
	c.Assert(err, IsNil)
	c.Assert(sent, Equals, 0)
}
