package statsite

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type ClientSuite struct {
	mockNetwork  Network
	mockStatsite *mockStatsite
}

var _ = Suite(&ClientSuite{})

func (s *ClientSuite) SetUpTest(c *C) {
	s.mockStatsite = &mockStatsite{}
	serverMap := make(map[string]mockServer)
	serverMap["statsite"] = mockServer(s.mockStatsite)
	s.mockNetwork = NewMockNetwork(serverMap)
}

func (s *ClientSuite) TestNewClient(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, IsNil)
	c.Assert(m.(*client).Conn, NotNil)
}

func (s *ClientSuite) TestConnectInvalidAddress(c *C) {
	m := NewNetworkClient("invalid", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, ErrorMatches, "Error resolving statsite:.*")
}

func (s *ClientSuite) TestNewClientConnect(c *C) {
	m := NewNetworkClient("badconnection", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, ErrorMatches, "Error connecting to statsite:.*")
}

func (s *ClientSuite) TestEmit(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, IsNil)

	msg := NewKeyValue("key", "value")
	// test 0 sent
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	err = m.Emit(msg)
	c.Assert(err, IsNil)
	// Expect 1 stat added to statsite
	c.Assert(s.mockStatsite.Count(), Equals, 1)
	c.Assert(s.mockStatsite.Last(), Equals, msg.String())
}

func (s *ClientSuite) TestEmitMultiple(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, IsNil)

	msg := NewKeyValue("key", "value")
	// test 0 sent
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	for i := 0; i < 10; i++ {
		err := m.Connect()
		c.Assert(err, IsNil)
		err = m.Emit(msg)
		c.Assert(err, IsNil)
	}
	// Expect 1 stat added to statsite
	c.Assert(s.mockStatsite.Count(), Equals, 10)
	c.Assert(s.mockStatsite.Last(), Equals, msg.String())
}

func (s *ClientSuite) TestEmitDouble(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, IsNil)

	msg := NewKeyValue("key", "value")
	// test 0 sent
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	err = m.Emit(msg)
	c.Assert(err, IsNil)
	// Expect 1 stat added to statsite
	c.Assert(s.mockStatsite.Count(), Equals, 1)
	m.Connect()
	m.Emit(msg)
	c.Assert(s.mockStatsite.Count(), Equals, 2)
}

func (s *ClientSuite) TestEmitNotConnected(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	conn := m.(*client).Conn
	c.Assert(conn, IsNil)
	msg := NewKeyValue("key", "value")
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	err := m.Emit(msg)
	c.Assert(err, IsNil)
	conn = m.(*client).Conn
	c.Assert(conn, NotNil)
	// Expect 1 stat added to statsite
	c.Assert(s.mockStatsite.Count(), Equals, 1)
}

func (s *ClientSuite) TestEmitFailure(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, IsNil)
	msg := NewKeyValue("bad", "key")
	// test 0 sent
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	// Expect error
	err = m.Emit(msg)
	c.Assert(err, NotNil)
	// Expect no stats added to statsite
	c.Assert(s.mockStatsite.Count(), Equals, 0)
}
