package basil_sshark

import (
	"bytes"
	"github.com/remogatto/prettytest"
	. "launchpad.net/gocheck"
	"testing"
)

type SSuite struct {
	prettytest.Suite
}

func TestStateRunner(t *testing.T) {
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(SSuite),
	)
}

func (s *SSuite) TestParsingState() {
	state, err := ParseState(bytes.NewBuffer([]byte(
		`{
			"id":"abc",
			"sessions": {
				"abc": {
					"port": 123,
					"container": "foocontainer"
				},
				"def": {
					"port": 456,
					"container": "anothercontainer"
				}
			}
		}`,
	)))

	s.Nil(err)

	s.Equal(state.ID, "abc")
	s.Check(state.Sessions, DeepEquals, map[string]Session{
		"abc": Session{Port: 123},
		"def": Session{Port: 456},
	})
}
