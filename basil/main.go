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

var ssharkState = flag.String("sshark", "/tmp/sshark.json", "path to sshark state file")

func main() {
	flag.Parse()

	mbus, err := go_cfmessagebus.NewCFMessageBus("NATS")
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	mbus.Configure(
		"127.0.0.1",
		4222,
		"",
		"",
	)

	err = mbus.Connect()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	registrator := basil.NewRegistrator("127.0.0.1", mbus)

	registrar := basil_sshark.NewRegistrar(registrator)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
		return
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Println("WHOA SHIT:", ev)

				if ev.IsModify() {
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
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Watch(*ssharkState)
	if err != nil {
		log.Fatal(err)
	}

	select {}

	watcher.Close()
}
