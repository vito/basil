package basil_sshark

import (
	"github.com/cloudfoundry/go_cfmessagebus/mock_cfmessagebus"
	"github.com/vito/basil"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"os"
	"time"
)

type WSuite struct {
	stateFile *os.File
}

func init() {
	Suite(&WSuite{})
}

func (s *WSuite) SetUpTest(c *C) {
	tmpdir := os.TempDir()

	file, err := ioutil.TempFile(tmpdir, "sshark-watcher-state-file")
	c.Assert(err, IsNil)

	err = ioutil.WriteFile(file.Name(), []byte(`{"id":"abc","sessions":{}}`), 0644)
	c.Assert(err, IsNil)

	s.stateFile = file
}

func (s *WSuite) TearDownTest(c *C) {
	err := os.Remove(s.stateFile.Name())
	c.Assert(err, IsNil)
}

func (s *WSuite) TestReactingToState(c *C) {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := mock_cfmessagebus.NewMockMessageBus()

	err := ReactTo(watcher, mbus, basil.DefaultConfig)
	c.Assert(err, IsNil)

	registered := make(chan []byte)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- msg
	})

	err = ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"abc","sessions":{"abc":{"port":123,"container":"foo"}}}`),
		0644,
	)

	c.Assert(err, IsNil)

	select {
	case msg := <-registered:
		c.Assert(string(msg), Equals, `{"uris":["abc"],"host":"127.0.0.1","port":123}`)
	case <-time.After(500 * time.Millisecond):
		c.Error("did not receive a router.register!")
	}
}

func (s *WSuite) TestHandlingInitialState(c *C) {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := mock_cfmessagebus.NewMockMessageBus()

	err := ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"abc","sessions":{"abc":{"port":123,"container":"foo"}}}`),
		0644,
	)
	c.Assert(err, IsNil)

	registered := make(chan []byte)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- msg
	})

	err = ReactTo(watcher, mbus, basil.DefaultConfig)
	c.Assert(err, IsNil)

	select {
	case msg := <-registered:
		c.Assert(string(msg), Equals, `{"uris":["abc"],"host":"127.0.0.1","port":123}`)
	case <-time.After(500 * time.Millisecond):
		c.Error("did not receive a router.register!")
	}
}

func (s *WSuite) TestReactingToRouterStart(c *C) {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := mock_cfmessagebus.NewMockMessageBus()

	err := ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"abc","sessions":{"abc":{"port":123,"container":"foo"}}}`),
		0644,
	)
	c.Assert(err, IsNil)

	err = ReactTo(watcher, mbus, basil.DefaultConfig)
	c.Assert(err, IsNil)

	registered := make(chan time.Time)

	mbus.Subscribe("router.register", func(msg []byte) {
		registered <- time.Now()
	})

	mbus.Publish("router.start", []byte(`{"minimumRegisterIntervalInSeconds":1}`))

	time1 := timedReceive(registered, 2*time.Second)
	c.Assert(time1, NotNil)

	time2 := timedReceive(registered, 2*time.Second)
	c.Assert(time2, NotNil)

	c.Assert((*time2).Sub(*time1) >= 1*time.Second, Equals, true)
}

func (s *WSuite) TestReactorSendsAdvertisements(c *C) {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := mock_cfmessagebus.NewMockMessageBus()

	err := ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"abc","sessions":{}}`),
		0644,
	)
	c.Assert(err, IsNil)

	config := basil.DefaultConfig
	config.AdvertiseInterval = 100 * time.Millisecond

	err = ReactTo(watcher, mbus, config)
	c.Assert(err, IsNil)

	advertised := make(chan time.Time)

	mbus.Subscribe("ssh.advertise", func(msg []byte) {
		advertised <- time.Now()
	})

	time1 := timedReceive(advertised, 1*time.Second)
	c.Assert(time1, NotNil)

	time2 := timedReceive(advertised, 1*time.Second)
	c.Assert(time2, NotNil)

	c.Assert((*time2).Sub(*time1) >= 100*time.Millisecond, Equals, true)
}

func (s *WSuite) TestReactorSendsAdvertisementsWithUpdatedID(c *C) {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := mock_cfmessagebus.NewMockMessageBus()

	err := ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"abc","sessions":{}}`),
		0644,
	)
	c.Assert(err, IsNil)

	config := basil.DefaultConfig
	config.AdvertiseInterval = 100 * time.Millisecond

	err = ReactTo(watcher, mbus, config)
	c.Assert(err, IsNil)

	advertised := make(chan []byte)

	mbus.Subscribe("ssh.advertise", func(msg []byte) {
		advertised <- msg
	})

	msg1 := waitReceive(advertised, 2*time.Second)
	c.Assert(string(msg1), Equals, `{"id":"abc"}`)

	err = ioutil.WriteFile(
		s.stateFile.Name(),
		[]byte(`{"id":"def","sessions":{}}`),
		0644,
	)
	c.Assert(err, IsNil)

	msg2 := waitReceive(advertised, 2*time.Second)
	c.Assert(string(msg2), Equals, `{"id":"def"}`)
}
