package basil

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
	"github.com/remogatto/prettytest"
	"testing"
	"errors"
)

type SWSuite struct {
	stateFile *os.File

	prettytest.Suite
}

func TestStateWatcherRunner(t *testing.T) {
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(SWSuite),
	)
}

func (s *SWSuite) Before() {
	tmpdir := os.TempDir()

	file, err := ioutil.TempFile(tmpdir, "state-watcher-state-file")
	s.Nil(err)

	s.stateFile = file
}

func (s *SWSuite) After() {
	err := os.Remove(s.stateFile.Name())
	s.Nil(err)
}

func (s *SWSuite) TestStateWatcherSeesModificationsAndInitialState() {
	sw := NewStateWatcher(s.stateFile.Name())

	modified := make(chan []byte)

	err := sw.OnStateChange(func(io io.Reader) {
		contents, err := ioutil.ReadAll(io)
		s.Nil(err)

		modified <- contents
	})
	s.Nil(err)

	val, err := waitReceive(modified)
	s.Nil(err)
	s.Equal(string(val), "")

	writeAbc := exec.Command("echo", "abc")
	writeAbc.Stdout = s.stateFile
	err = writeAbc.Run()
	s.Nil(err)

	val, err = waitReceive(modified)
	s.Nil(err)
	s.Equal(string(val), "abc\n")

	writeDef := exec.Command("echo", "def")
	writeDef.Stdout = s.stateFile
	err = writeDef.Run()
	s.Nil(err)

	val, err = waitReceive(modified)
	s.Nil(err)
	s.Equal(string(val), "abc\ndef\n")
}

func waitReceive(from chan []byte) ([]byte, error) {
	select {
	case val := <-from:
		return val, nil
	case <-time.After(500 * time.Millisecond):
		return nil, errors.New("timeout")
	}
}
