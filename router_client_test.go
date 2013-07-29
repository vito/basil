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

func (s *RSuite) TestRouterClientRegistering(c *C) {
	mbus := go_cfmessagebus.NewMockMessageBus()

	routerClient := NewRouterClient("1.2.3.4", mbus)

	registered := make(chan []byte)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- msg
	})

	routerClient.Register(123, []string{"abc"})

	select {
	case msg := <-registered:
		c.Assert(string(msg), Equals, `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		c.Error("did not receive a router.register!")
	}
}

func (s *RSuite) TestRouterClientUnregistering(c *C) {
	mbus := go_cfmessagebus.NewMockMessageBus()

	routerClient := NewRouterClient("1.2.3.4", mbus)

	registered := make(chan []byte)

	mbus.Subscribe("router.unregister", func(msg []byte) {
		registered <- msg
	})

	routerClient.Unregister(123, []string{"abc"})

	select {
	case msg := <-registered:
		c.Assert(string(msg), Equals, `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		c.Error("did not receive a router.unregister!")
	}
}

func (s *RSuite) TestRouterClientRouterStartHandling(c *C) {
	mbus := go_cfmessagebus.NewMockMessageBus()

	routerClient := NewRouterClient("1.2.3.4", mbus)

	times := make(chan time.Time)

	routerClient.Periodically(func() {
		times <- time.Now()
	})

	err := routerClient.Greet()
	c.Assert(err, IsNil)

	mbus.Publish("router.start", []byte(`{"minimumRegisterIntervalInSeconds":1}`))

	time1 := timedReceive(times, 2*time.Second)
	c.Assert(time1, NotNil)

	time2 := timedReceive(times, 2*time.Second)
	c.Assert(time2, NotNil)

	c.Assert((*time2).Sub(*time1) >= 1*time.Second, Equals, true)
}

func (s *RSuite) TestRouterClientGreeting(c *C) {
	mbus := go_cfmessagebus.NewMockMessageBus()

	routerClient := NewRouterClient("1.2.3.4", mbus)

	times := make(chan time.Time)

	routerClient.Periodically(func() {
		times <- time.Now()
	})

	mbus.RespondToChannel("router.greet", func([]byte) []byte {
		return []byte(`{"minimumRegisterIntervalInSeconds":1}`)
	})

	err := routerClient.Greet()
	c.Assert(err, IsNil)

	time1 := timedReceive(times, 2*time.Second)
	c.Assert(time1, NotNil)

	time2 := timedReceive(times, 2*time.Second)
	c.Assert(time2, NotNil)

	c.Assert((*time2).Sub(*time1) >= 1*time.Second, Equals, true)
}

func timedReceive(from chan time.Time, giveup time.Duration) *time.Time {
	select {
	case val := <-from:
		return &val
	case <-time.After(giveup):
		return nil
	}
}
