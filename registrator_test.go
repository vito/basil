package basil

import (
	"github.com/cloudfoundry/go_cfmessagebus"
	. "launchpad.net/gocheck"
	"time"
)

type RSuite struct{}

func init() {
	Suite(&RSuite{})
}

func (s *RSuite) TestRegistratorRegistering(c *C) {
	mbus := go_cfmessagebus.NewMockMessageBus()

	registrator := NewRegistrator("1.2.3.4", mbus)

	registered := make(chan []byte)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- msg
	})

	registrator.Register(123, []string{"abc"})

	select {
	case msg := <-registered:
		c.Assert(string(msg), Equals, `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		c.Error("did not receive a router.register!")
	}
}

func (s *RSuite) TestRegistratorUnregistering(c *C) {
	mbus := go_cfmessagebus.NewMockMessageBus()

	registrator := NewRegistrator("1.2.3.4", mbus)

	registered := make(chan []byte)

	mbus.Subscribe("router.unregister", func(msg []byte) {
		registered <- msg
	})

	registrator.Unregister(123, []string{"abc"})

	select {
	case msg := <-registered:
		c.Assert(string(msg), Equals, `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		c.Error("did not receive a router.unregister!")
	}
}
