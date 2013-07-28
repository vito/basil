package basil_sshark

import (
	"encoding/json"
	"github.com/cloudfoundry/sshark"
)

type Session struct {
	Port sshark.MappedPort `json:"port"`
}

type State struct {
	ID       string             `json:"id"`
	Sessions map[string]Session `json:"sessions"`
}

func ParseState(stateJSON []byte) (*State, error) {
	state := &State{}

	err := json.Unmarshal(stateJSON, state)
	if err != nil {
		return nil, err
	}

	return state, nil
}
