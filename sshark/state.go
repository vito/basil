package basil_sshark

import (
	"encoding/json"
	"github.com/cloudfoundry/sshark"
	"io"
)

type Session struct {
	Port sshark.MappedPort `json:"port"`
}

type State struct {
	ID       string             `json:"id"`
	Sessions map[string]Session `json:"sessions"`
}

func ParseState(stateIO io.Reader) (*State, error) {
	state := &State{}

	decoder := json.NewDecoder(stateIO)
	err := decoder.Decode(state)
	if err != nil {
		return nil, err
	}

	return state, nil
}
