package basil_sshark

import (
	"bytes"
	. "launchpad.net/gocheck"
)

type SSuite struct{}

func init() {
	Suite(&SSuite{})
}

func (s *SSuite) TestParsingState(c *C) {
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

	c.Assert(err, IsNil)

	c.Assert(state.ID, Equals, "abc")
	c.Assert(state.Sessions, DeepEquals, map[string]Session{
		"abc": Session{Port: 123},
		"def": Session{Port: 456},
	})
}
