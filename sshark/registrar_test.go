package basil_sshark

import (
	"github.com/cloudfoundry/go_cfmessagebus/mock_cfmessagebus"
	"github.com/remogatto/prettytest"
	"github.com/vito/basil"
	"testing"
	"time"
)

type SRSuite struct {
	prettytest.Suite
}

func TestRegistrarRunner(t *testing.T) {
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(SRSuite),
	)
}

func (s *SRSuite) TestRegistrarUpdateRegisters() {
	mbus := mock_cfmessagebus.NewMockMessageBus()

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
		s.Equal(string(msg), `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		s.Error("did not receive a router.register!")
	}
}

func (s *SRSuite) TestRegistrarUpdateUnregisters() {
	mbus := mock_cfmessagebus.NewMockMessageBus()

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
		s.Equal(string(msg), `{"uris":["abc"],"host":"1.2.3.4","port":123}`)
	case <-time.After(500 * time.Millisecond):
		s.Error("did not receive a router.register!")
	}
}

func (s *SRSuite) TestPeriodicRegistration() {
	mbus := mock_cfmessagebus.NewMockMessageBus()

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

	err := registrar.RegisterPeriodically()
	s.Nil(err)

	time1 := timedReceive(registered, 2*time.Second)
	s.Not(s.Nil(time1))

	time2 := timedReceive(registered, 2*time.Second)
	s.Not(s.Nil(time2))

	s.True((*time2).Sub(*time1) >= 1*time.Second)
}
