package basil_sshark

import (
	"github.com/cloudfoundry/sshark"
	"github.com/vito/basil"
)

type Routes map[string]sshark.MappedPort

type Registrar struct {
	routes      Routes
	registrator *basil.Registrator
}

func NewRegistrar(registrator *basil.Registrator) *Registrar {
	return &Registrar{
		routes:      make(Routes),
		registrator: registrator,
	}
}

func (r *Registrar) Update(state *State) error {
	routes := make(Routes)
	for id, session := range state.Sessions {
		routes[id] = session.Port
	}

	newRoutes := routesDiff(routes, r.routes)
	oldRoutes := routesDiff(r.routes, routes)

	r.registerRoutes(newRoutes)
	r.unregisterRoutes(oldRoutes)

	r.routes = routes

	return nil
}

func (r *Registrar) registerRoutes(routes Routes) {
	for id, port := range routes {
		r.registrator.Register(int(port), []string{id})
	}
}

func (r *Registrar) unregisterRoutes(routes Routes) {
	for id, port := range routes {
		r.registrator.Unregister(int(port), []string{id})
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
