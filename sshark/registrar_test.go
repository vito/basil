package basil_sshark

import (
	"github.com/cloudfoundry/go_cfmessagebus"
	"github.com/vito/basil"
	. "launchpad.net/gocheck"
	"time"
)

type SRSuite struct{}

func init() {
	Suite(&SRSuite{})
}

func (s *SRSuite) TestRegistrarUpdateRegisters(c *C) {
	mbus := go_cfmessagebus.NewMockMessageBus()

	routerClient := basil.NewRouterClient("1.2.3.4", mbus)

	registered := make(chan []byte)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- msg
	})

	registrar := NewRegistrar(routerClient)
	registrar.Update(&State{
		Sessions: map[string]Session{
			"abc": Session{
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

	routerClient := basil.NewRouterClient("1.2.3.4", mbus)

	unregistered := make(chan []byte)

	mbus.Subscribe("router.unregister", func(msg []byte) {
		unregistered <- msg
	})

	registrar := NewRegistrar(routerClient)
	registrar.Update(&State{
		ID: "foo",
		Sessions: map[string]Session{
			"abc": Session{
				Port: 123,
			},
		},
	})

	registrar.Update(&State{
		ID:       "foo",
		Sessions: map[string]Session{},
	})

	select {
	case msg := <-unregistered:
		c.Assert(string(msg), Equals, `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		c.Error("did not receive a router.register!")
	}
}

func (s *SRSuite) TestPeriodicRegistration(c *C) {
	mbus := go_cfmessagebus.NewMockMessageBus()

	routerClient := basil.NewRouterClient("1.2.3.4", mbus)

	registrar := NewRegistrar(routerClient)
	registrar.Update(&State{
		ID: "foo",
		Sessions: map[string]Session{
			"abc": Session{
				Port: 123,
			},
		},
	})

	registered := make(chan time.Time)
	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- time.Now()
	})

	mbus.RespondToChannel("router.greet", func([]byte) []byte {
		return []byte(`{"minimumRegisterIntervalInSeconds":1}`)
	})

	err := registrar.PeriodicallyRegister()
	c.Assert(err, IsNil)

	time1 := timedReceive(registered, 2*time.Second)
	c.Assert(time1, NotNil)

	time2 := timedReceive(registered, 2*time.Second)
	c.Assert(time2, NotNil)

	c.Assert((*time2).Sub(*time1) >= 1*time.Second, Equals, true)
}
