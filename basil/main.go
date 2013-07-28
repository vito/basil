package main

import (
	"flag"
	"github.com/cloudfoundry/go_cfmessagebus"
	"github.com/howeyc/fsnotify"
	"github.com/vito/basil"
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

	registrar := basil.NewSSHarkRegistrar(mbus)

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
					log.Println("updating!")
					body, err := ioutil.ReadFile(*ssharkState)
					if err == nil {
						log.Println("sup")
						registrar.Update(body)
					}
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
