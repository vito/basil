package basil

import (
	"encoding/json"
	"github.com/cloudfoundry/go_cfmessagebus"
	"log"
	"time"
)

type RouterClient struct {
	Host       string
	messageBus cfmessagebus.MessageBus

	periodicCallback     func()
	stopPeriodicCallback chan bool
}

type RegistryMessage struct {
	URIs []string `json:"uris"`
	Host string   `json:"host"`
	Port int      `json:"port"`
}

type RouterGreetingMessage struct {
	MinimumRegisterInterval int `json:"minimumRegisterIntervalInSeconds"`
}

func NewRouterClient(host string, messageBus cfmessagebus.MessageBus) *RouterClient {
	return &RouterClient{
		Host:       host,
		messageBus: messageBus,
	}
}

func (r *RouterClient) Periodically(callback func()) {
	r.periodicCallback = callback
}

func (r *RouterClient) Greet() error {
	err := r.messageBus.Subscribe("router.start", r.handleGreeting)
	if err != nil {
		return err
	}

	return r.messageBus.Request("router.greet", []byte{}, r.handleGreeting)
}

func (r *RouterClient) Register(port int, uris []string) error {
	return r.sendRegistryMessage("router.register", port, uris)
}

func (r *RouterClient) Unregister(port int, uris []string) error {
	return r.sendRegistryMessage("router.unregister", port, uris)
}

func (r *RouterClient) handleGreeting(greeting []byte) {
	interval, err := r.intervalFrom(greeting)
	if err != nil {
		log.Printf("failed to parse router.start: %s\n", err)
		return
	}

	go r.callbackPeriodically(time.Duration(interval) * time.Second)
}

func (r *RouterClient) callbackPeriodically(interval time.Duration) {
	if r.stopPeriodicCallback != nil {
		r.stopPeriodicCallback <- true
	}

	callback := r.periodicCallback

	if callback == nil {
		return
	}

	cancel := make(chan bool)

	r.stopPeriodicCallback = cancel

	for stop := false; !stop; {
		select {
		case <-time.After(interval):
			callback()
		case stop = <-cancel:
		}
	}
}

func (r *RouterClient) sendRegistryMessage(subject string, port int, uris []string) error {
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

func (r *RouterClient) intervalFrom(greetingPayload []byte) (int, error) {
	var greeting RouterGreetingMessage

	err := json.Unmarshal(greetingPayload, &greeting)
	if err != nil {
		return 0, err
	}

	return greeting.MinimumRegisterInterval, nil
}
