package basil_sshark

import (
	"github.com/cloudfoundry/go_cfmessagebus"
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

func (s *WSuite) SetUpSuite(c *C) {
	tmpdir := os.TempDir()

	file, err := ioutil.TempFile(tmpdir, "sshark-watcher-state-file")
	c.Assert(err, IsNil)

	err = ioutil.WriteFile(file.Name(), []byte(`{"id":"abc","sessions":{}}`), 0644)
	c.Assert(err, IsNil)

	s.stateFile = file
}

func (s *WSuite) TearDownSuite(c *C) {
	err := os.Remove(s.stateFile.Name())
	c.Assert(err, IsNil)
}

func (s *WSuite) TestReactingToState(c *C) {
	watcher := basil.NewStateWatcher(s.stateFile.Name())

	mbus := go_cfmessagebus.NewMockMessageBus()

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
