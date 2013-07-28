package basil

import (
	"encoding/json"
	"github.com/cloudfoundry/go_cfmessagebus"
)

type Registrator struct {
	Host       string
	messageBus go_cfmessagebus.CFMessageBus
}

type RegistryMessage struct {
	URIs []string `json:"uris"`
	Host string   `json:"host"`
	Port int      `json:"port"`
}

func NewRegistrator(host string, messageBus go_cfmessagebus.CFMessageBus) *Registrator {
	return &Registrator{
		Host:       host,
		messageBus: messageBus,
	}
}

func (r *Registrator) Register(port int, uris []string) error {
	return r.sendRegistryMessage("router.register", port, uris)
}

func (r *Registrator) Unregister(port int, uris []string) error {
	return r.sendRegistryMessage("router.unregister", port, uris)
}

func (r *Registrator) sendRegistryMessage(subject string, port int, uris []string) error {
	msg := &RegistryMessage{
		URIs: uris,
		Host: r.Host,
		Port: port,
	}

	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return r.messageBus.Publish(subject, json)
}
