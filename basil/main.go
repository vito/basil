package main

import (
	"flag"
	"github.com/cloudfoundry/go_cfmessagebus"
	"github.com/howeyc/fsnotify"
	"github.com/vito/basil"
	"github.com/vito/basil/sshark"
	"io/ioutil"
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

	registrator := basil.NewRegistrator(config.Host, mbus)

	registrar := basil_sshark.NewRegistrar(registrator)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
		return
	}

	go func() {
		for {
			select {
			case <-watcher.Event:
				body, err := ioutil.ReadFile(*ssharkState)
				if err != nil {
					log.Printf("failed to read sshark state: %s\n", err)
					break
				}

				state, err := basil_sshark.ParseState(body)
				if err != nil {
					log.Printf("failed to parse sshark state: %s\n", err)
					break
				}

				registrar.Update(state)
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.WatchFlags(*ssharkState, fsnotify.FSN_MODIFY)
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
