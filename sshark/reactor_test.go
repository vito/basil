package basil_sshark

import (
	"github.com/cloudfoundry/go_cfmessagebus/mock_cfmessagebus"
	"github.com/remogatto/prettytest"
	"github.com/vito/basil"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type WSuite struct {
	stateFile *os.File

	prettytest.Suite
}

func TestReactorRunner(t *testing.T) {
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(WSuite),
	)
}

func (s *WSuite) Before() {
	tmpdir := os.TempDir()

	file, err := ioutil.TempFile(tmpdir, "sshark-watcher-state-file")
	s.Nil(err)

	err = ioutil.WriteFile(file.Name(), []byte(`{"id":"abc","sessions":{}}`), 0644)
	s.Nil(err)

	s.stateFile = file
}

func (s *WSuite) After() {
	err := os.Remove(s.stateFile.Name())
	s.Nil(err)
}

func (s *WSuite) TestReactingToState() {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := mock_cfmessagebus.NewMockMessageBus()

	err := ReactTo(watcher, mbus, basil.DefaultConfig)
	s.Nil(err)

	registered := make(chan []byte)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- msg
	})

	err = ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"abc","sessions":{"abc":{"port":123,"container":"foo"}}}`),
		0644,
	)

	s.Nil(err)

	select {
	case msg := <-registered:
		s.Equal(string(msg), `{"uris":["abc"],"host":"127.0.0.1","port":123}`)
	case <-time.After(500 * time.Millisecond):
		s.Error("did not receive a router.register!")
	}
}

func (s *WSuite) TestHandlingInitialState() {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := mock_cfmessagebus.NewMockMessageBus()

	err := ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"abc","sessions":{"abc":{"port":123,"container":"foo"}}}`),
		0644,
	)
	s.Nil(err)

	registered := make(chan []byte)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- msg
	})

	err = ReactTo(watcher, mbus, basil.DefaultConfig)
	s.Nil(err)

	select {
	case msg := <-registered:
		s.Equal(string(msg), `{"uris":["abc"],"host":"127.0.0.1","port":123}`)
	case <-time.After(500 * time.Millisecond):
		s.Error("did not receive a router.register!")
	}
}

func (s *WSuite) TestReactingToRouterStart() {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := mock_cfmessagebus.NewMockMessageBus()

	err := ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"abc","sessions":{"abc":{"port":123,"container":"foo"}}}`),
		0644,
	)
	s.Nil(err)

	err = ReactTo(watcher, mbus, basil.DefaultConfig)
	s.Nil(err)

	registered := make(chan time.Time)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- time.Now()
	})

	mbus.Publish("router.start", []byte(`{"minimumRegisterIntervalInSeconds":1}`))

	time1 := timedReceive(registered, 2*time.Second)
	s.Not(s.Nil(time1))

	time2 := timedReceive(registered, 2*time.Second)
	s.Not(s.Nil(time2))

	s.True((*time2).Sub(*time1) >= 1*time.Second)
}

func (s *WSuite) TestReactorSendsAdvertisements() {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := mock_cfmessagebus.NewMockMessageBus()

	err := ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"abc","sessions":{}}`),
		0644,
	)
	s.Nil(err)

	config := basil.DefaultConfig
	config.AdvertiseInterval = 100 * time.Millisecond

	err = ReactTo(watcher, mbus, config)
	s.Nil(err)

	advertised := make(chan time.Time)

	mbus.Subscribe("ssh.advertise", func(msg []byte) {
		advertised <- time.Now()
	})

	time1 := timedReceive(advertised, 1*time.Second)
	s.Not(s.Nil(time1))

	time2 := timedReceive(advertised, 1*time.Second)
	s.Not(s.Nil(time2))

	s.True((*time2).Sub(*time1) >= 100*time.Millisecond)
}

func (s *WSuite) TestReactorSendsAdvertisementsWithUpdatedID() {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := mock_cfmessagebus.NewMockMessageBus()

	err := ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"abc","sessions":{}}`),
		0644,
	)
	s.Nil(err)

	config := basil.DefaultConfig
	config.AdvertiseInterval = 100 * time.Millisecond

	err = ReactTo(watcher, mbus, config)
	s.Nil(err)

	advertised := make(chan []byte)

	mbus.Subscribe("ssh.advertise", func(msg []byte) {
		advertised <- msg
	})

	msg1 := waitReceive(advertised, 2*time.Second)
	s.Equal(string(msg1), `{"id":"abc"}`)

	err = ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"def","sessions":{}}`),
		0644,
	)
	s.Nil(err)

	msg2 := waitReceive(advertised, 2*time.Second)
	s.Equal(string(msg2), `{"id":"def"}`)
}
