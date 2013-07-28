package basil

import (
	"github.com/cloudfoundry/go_cfmessagebus"
	. "launchpad.net/gocheck"
	"time"
)

type SRSuite struct{}

func init() {
	Suite(&SRSuite{})
}

func (s *SRSuite) TestRegistrarUpdateRegisters(c *C) {
	mbus := go_cfmessagebus.NewMockMessageBus()

	registered := make(chan []byte)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- msg
	})

	registrar := NewSSHarkRegistrar(mbus)
	registrar.Update(&SSHarkState{
		Sessions: map[string]SSHarkSession{
			"abc": SSHarkSession{
				Port: 123,
			},
		},
	})

	select {
	case msg := <-registered:
		c.Assert(string(msg), Equals, `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		c.Error("did not receive a router.register!")
	}
}

func (s *SRSuite) TestRegistrarUpdateUnregisters(c *C) {
	mbus := go_cfmessagebus.NewMockMessageBus()

	unregistered := make(chan []byte)

	mbus.Subscribe("router.unregister", func(msg []byte) {
		unregistered <- msg
	})

	registrar := NewSSHarkRegistrar(mbus)
	registrar.Update(&SSHarkState{
		ID: "foo",
		Sessions: map[string]SSHarkSession{
			"abc": SSHarkSession{
				Port: 123,
			},
		},
	})

	registrar.Update(&SSHarkState{
		ID:       "foo",
		Sessions: map[string]SSHarkSession{},
	})

	select {
	case msg := <-unregistered:
		c.Assert(string(msg), Equals, `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		c.Error("did not receive a router.register!")
	}
}
