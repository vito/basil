package basil

import (
	"log"
	"encoding/json"
	"github.com/cloudfoundry/go_cfmessagebus"
	"github.com/cloudfoundry/sshark"
)

type Routes map[string]sshark.MappedPort

type SSHarkRegistrar struct {
	routes Routes
	messageBus go_cfmessagebus.CFMessageBus
}

func NewSSHarkRegistrar(mbus go_cfmessagebus.CFMessageBus) *SSHarkRegistrar {
	return &SSHarkRegistrar{
		routes: make(Routes),
		messageBus: mbus,
	}
}

type RegistryMessage struct {
	URIs []string `json:"uris"`
	Host string `json:"host"`
	Port sshark.MappedPort `json:"port"`
}

func (r *SSHarkRegistrar) Update(state *SSHarkState) error {
	routes := make(Routes)
	for id, session := range state.Sessions {
		routes[id] = session.Port
	}

	newRoutes := routesDiff(routes, r.routes)
	oldRoutes := routesDiff(r.routes, routes)

	go r.registerRoutes(newRoutes)
	go r.unregisterRoutes(oldRoutes)

	r.routes = routes

	return nil
}

func (r *SSHarkRegistrar) registerRoutes(routes Routes) {
	r.sendRegistryMessage("router.register", routes)
}

func (r *SSHarkRegistrar) unregisterRoutes(routes Routes) {
	r.sendRegistryMessage("router.unregister", routes)
}

func (r *SSHarkRegistrar) sendRegistryMessage(subject string, routes Routes) {
	for id, port := range routes {
		msg := &RegistryMessage{
			URIs: []string{id},
			Host: "1.2.3.4", // TODO
			Port: port,
		}

		json, err := json.Marshal(msg)
		if err != nil {
			log.Printf("failed to marshal: %s\n", err)
			continue
		}

		r.messageBus.Publish(subject, json)
	}
}

func routesDiff(a, b Routes) Routes {
	routes := make(Routes)

	for id, port := range a {
		_, present := b[id]
		if !present {
			routes[id] = port
		}
	}

	return routes
}
