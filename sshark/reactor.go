package basil_sshark

import (
	"github.com/cloudfoundry/go_cfmessagebus"
	"github.com/vito/basil"
	"io"
	"log"
)

func ReactTo(watcher *basil.StateWatcher, mbus go_cfmessagebus.CFMessageBus, config basil.Config) {
	registrator := basil.NewRegistrator(config.Host, mbus)

	registrar := NewRegistrar(registrator)

	watcher.OnModify(func(body io.Reader) {
		state, err := ParseState(body)
		if err != nil {
			log.Printf("failed to parse sshark state: %s\n", err)
			return
		}

		registrar.Update(state)
	})
}
