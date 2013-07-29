package basil_sshark

import (
	"github.com/cloudfoundry/sshark"
	"github.com/vito/basil"
	"sync"
)

type Routes map[string]sshark.MappedPort

type Registrar struct {
	routes       Routes
	routerClient *basil.RouterClient

	stopPeriodicRegistration chan bool

	lock sync.RWMutex
}

func NewRegistrar(routerClient *basil.RouterClient) *Registrar {
	return &Registrar{
		routes:       make(Routes),
		routerClient: routerClient,
	}
}

func (r *Registrar) Update(state *State) error {
	routes := make(Routes)
	for id, session := range state.Sessions {
		routes[id] = session.Port
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	newRoutes := routesDiff(routes, r.routes)
	oldRoutes := routesDiff(r.routes, routes)

	r.registerRoutes(newRoutes)
	r.unregisterRoutes(oldRoutes)

	r.routes = routes

	return nil
}

func (r *Registrar) BroadcastCurrentRoutes() {
	r.lock.RLock()
	defer r.lock.RUnlock()

	r.registerRoutes(r.routes)
}

func (r *Registrar) PeriodicallyRegister() error {
	r.routerClient.Periodically(r.BroadcastCurrentRoutes)

	return r.routerClient.Greet()
}

func (r *Registrar) registerRoutes(routes Routes) {
	for id, port := range routes {
		r.routerClient.Register(int(port), []string{id})
	}
}

func (r *Registrar) unregisterRoutes(routes Routes) {
	for id, port := range routes {
		r.routerClient.Unregister(int(port), []string{id})
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
