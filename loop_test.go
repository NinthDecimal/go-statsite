package statsite

import (
	"fmt"

	. "gopkg.in/check.v1"
)

// func Test(t *testing.T) {
// 	TestingT(t)
// }

type LoopSuite struct {
	client       Client
	mockNetwork  Network
	mockStatsite *mockStatsite
}

var _ = Suite(&LoopSuite{})

func (s *LoopSuite) SetUpTest(c *C) {
	s.mockStatsite = &mockStatsite{}
	serverMap := make(map[string]mockServer)
	serverMap["statsite"] = mockServer(s.mockStatsite)
	s.mockNetwork = NewMockNetwork(serverMap)
	s.client = NewNetworkClient("statsite", s.mockNetwork)
}

func (s *LoopSuite) TestInitialize(c *C) {
	InitializeWithClient("foo.bar", s.client)
	c.Assert(enabled, Equals, true)
	Shutdown()
	c.Assert(enabled, Equals, false)
}

func (s *LoopSuite) TestFlushKV(c *C) {
	InitializeWithClient("foo.bar", s.client)
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	c.Assert(enabled, Equals, true)
	kv := KeyValue("loop", "test")
	kv.Emit()
	Shutdown()
	c.Assert(s.mockStatsite.Count(), Equals, 1)
	c.Assert(enabled, Equals, false)
}

func (s *LoopSuite) TestFlushKVMultiple(c *C) {
	InitializeWithClient("foo.bar", s.client)
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	c.Assert(enabled, Equals, true)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		kv := KeyValue(key, "value")
		kv.Emit()
	}
	Shutdown()
	c.Assert(s.mockStatsite.Count(), Equals, 10)
	c.Assert(enabled, Equals, false)
}

func (s *LoopSuite) TestShutdownNotInitialized(c *C) {
	c.Assert(enabled, Equals, false)
	Shutdown()
	c.Assert(enabled, Equals, false)
}

func _DeferFlush() {
	timer := Timer("foo")
	defer timer.Emit()
}

func (s *LoopSuite) TestFlushDefer(c *C) {
	InitializeWithClient("foo.bar", s.client)
	c.Assert(enabled, Equals, true)
	_DeferFlush()
	Shutdown()
	c.Assert(s.mockStatsite.Count(), Equals, 1)
	c.Assert(enabled, Equals, false)
}
