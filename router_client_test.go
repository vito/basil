package basil

import (
	"github.com/cloudfoundry/go_cfmessagebus/mock_cfmessagebus"
	"time"
	"github.com/remogatto/prettytest"
	"testing"
)

type RSuite struct {
	prettytest.Suite
}

func TestRouterClientRunner(t *testing.T) {
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(RSuite),
	)
}

func (s *RSuite) TestRouterClientRegistering() {
	mbus := mock_cfmessagebus.NewMockMessageBus()

	routerClient := NewRouterClient("1.2.3.4", mbus)

	registered := make(chan []byte)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- msg
	})

	routerClient.Register(123, []string{"abc"})

	select {
	case msg := <-registered:
		s.Equal(string(msg), `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		s.Error("did not receive a router.register!")
	}
}

func (s *RSuite) TestRouterClientUnregistering() {
	mbus := mock_cfmessagebus.NewMockMessageBus()

	routerClient := NewRouterClient("1.2.3.4", mbus)

	registered := make(chan []byte)

	mbus.Subscribe("router.unregister", func(msg []byte) {
		registered <- msg
	})

	routerClient.Unregister(123, []string{"abc"})

	select {
	case msg := <-registered:
		s.Equal(string(msg), `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		s.Error("did not receive a router.unregister!")
	}
}

func (s *RSuite) TestRouterClientRouterStartHandling() {
	mbus := mock_cfmessagebus.NewMockMessageBus()

	routerClient := NewRouterClient("1.2.3.4", mbus)

	times := make(chan time.Time)

	routerClient.Periodically(func() {
		times <- time.Now()
	})

	err := routerClient.Greet()
	s.Nil(err)

	mbus.Publish("router.start", []byte(`{"minimumRegisterIntervalInSeconds":1}`))

	time1 := timedReceive(times, 2*time.Second)
	s.Not(s.Nil(time1))

	time2 := timedReceive(times, 2*time.Second)
	s.Not(s.Nil(time2))

	s.True((*time2).Sub(*time1) >= 1*time.Second)
}

func (s *RSuite) TestRouterClientGreeting() {
	mbus := mock_cfmessagebus.NewMockMessageBus()

	routerClient := NewRouterClient("1.2.3.4", mbus)

	times := make(chan time.Time)

	routerClient.Periodically(func() {
		times <- time.Now()
	})

	mbus.RespondToChannel("router.greet", func([]byte) []byte {
		return []byte(`{"minimumRegisterIntervalInSeconds":1}`)
	})

	err := routerClient.Greet()
	s.Nil(err)

	time1 := timedReceive(times, 2*time.Second)
	s.Not(s.Nil(time1))

	time2 := timedReceive(times, 2*time.Second)
	s.Not(s.Nil(time2))

	s.True((*time2).Sub(*time1) >= 1*time.Second)
}

func timedReceive(from chan time.Time, giveup time.Duration) *time.Time {
	select {
	case val := <-from:
		return &val
	case <-time.After(giveup):
		return nil
	}
}
