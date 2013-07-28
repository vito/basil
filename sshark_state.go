package basil

import (
	"encoding/json"
	"github.com/cloudfoundry/sshark"
)

type SSHarkSession struct {
	Port sshark.MappedPort `json:"port"`
}

type SSHarkState struct {
	ID       string                   `json:"id"`
	Sessions map[string]SSHarkSession `json:"sessions"`
}

func ParseSSHarkState(stateJSON []byte) (*SSHarkState, error) {
	state := &SSHarkState{}

	err := json.Unmarshal(stateJSON, state)
	if err != nil {
		return nil, err
	}

	return state, nil
}
