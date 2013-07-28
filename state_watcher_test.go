package basil

import (
	"io"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"os"
	"os/exec"
	"time"
)

type SWSuite struct {
	stateFile *os.File
}

func init() {
	Suite(&SWSuite{})
}

func (s *SWSuite) SetUpSuite(c *C) {
	tmpdir := os.TempDir()

	file, err := ioutil.TempFile(tmpdir, "state-watcher-state-file")
	c.Assert(err, IsNil)

	s.stateFile = file
}

func (s *SWSuite) TearDownSuite(c *C) {
	err := os.Remove(s.stateFile.Name())
	c.Assert(err, IsNil)
}

func (s *SWSuite) TestStateWatcherSeesModifications(c *C) {
	sw := NewStateWatcher(s.stateFile.Name())

	modified := make(chan []byte)

	sw.OnModify(func(io io.Reader) {
		contents, err := ioutil.ReadAll(io)
		c.Assert(err, IsNil)

		modified <- contents
	})

	writeAbc := exec.Command("echo", "abc")
	writeAbc.Stdout = s.stateFile
	err := writeAbc.Run()
	c.Assert(err, IsNil)

	val := waitReceive(modified)
	c.Assert(string(val), Equals, "abc\n")

	writeDef := exec.Command("echo", "def")
	writeDef.Stdout = s.stateFile
	err = writeDef.Run()
	c.Assert(err, IsNil)

	val = waitReceive(modified)
	c.Assert(string(val), Equals, "abc\ndef\n")
}

func waitReceive(from chan []byte) []byte {
	select {
	case val := <-from:
		return val
	case <-time.After(500 * time.Millisecond):
		return nil
	}
}
