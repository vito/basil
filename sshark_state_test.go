package basil

import (
	. "launchpad.net/gocheck"
)

type SSSuite struct{}

func init() {
	Suite(&SSSuite{})
}

func (s *SSSuite) TestParsingSSHarkState(c *C) {
	state, err := ParseSSHarkState([]byte(
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
	))

	c.Assert(err, IsNil)

	c.Assert(state.ID, Equals, "abc")
	c.Assert(state.Sessions, DeepEquals, map[string]SSHarkSession{
		"abc": SSHarkSession{Port: 123},
		"def": SSHarkSession{Port: 456},
	})
}
