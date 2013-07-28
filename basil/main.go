package main

import (
	"flag"
	"github.com/cloudfoundry/go_cfmessagebus"
	"github.com/vito/basil"
	"github.com/vito/basil/sshark"
	"log"
)

var configFile = flag.String("config", "", "path to config file")

var ssharkState = flag.String(
	"sshark",
	"/tmp/sshark.json",
	"path to sshark state file",
)

func main() {
	flag.Parse()

	var config basil.Config

	if *configFile != "" {
		config = basil.LoadConfig(*configFile)
	} else {
		config = basil.DefaultConfig
	}

	mbus, err := go_cfmessagebus.NewCFMessageBus("NATS")
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	mbus.Configure(
		config.MessageBus.Host,
		config.MessageBus.Port,
		config.MessageBus.Username,
		config.MessageBus.Password,
	)

	err = mbus.Connect()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	watcher := basil.NewStateWatcher(*ssharkState)

	registrator := basil.NewRegistrator(config.Host, mbus)

	registrar := basil_sshark.NewRegistrar(registrator)

	err = watcher.OnModify(func(body []byte) {
		log.Printf("I see you have modified %s, sir!\n", *ssharkState)

		state, err := basil_sshark.ParseState(body)
		if err != nil {
			log.Printf("failed to parse sshark state: %s\n", err)
			return
		}

		registrar.Update(state)
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	select {}
}
