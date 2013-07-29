package basil_sshark

import (
	cf "github.com/cloudfoundry/go_cfmessagebus"
	"github.com/vito/basil"
	"io"
	"log"
)

func ReactTo(watcher *basil.StateWatcher, mbus cf.MessageBus, config basil.Config) error {
	routerClient := basil.NewRouterClient(config.Host, mbus)

	registrar := NewRegistrar(routerClient)

	err := registrar.PeriodicallyRegister()
	if err != nil {
		return err
	}

	return watcher.OnStateChange(func(body io.Reader) {
		state, err := ParseState(body)
		if err != nil {
			log.Printf("failed to parse sshark state: %s\n", err)
			return
		}

		registrar.Update(state)
	})
}
