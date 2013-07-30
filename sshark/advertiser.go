package basil_sshark

import (
	"encoding/json"
	"github.com/cloudfoundry/go_cfmessagebus"
	"github.com/vito/basil"
	"sync"
	"time"
)

type Advertiser struct {
	config basil.Config
	state  *State

	sync.RWMutex
}

type AdvertiseMessage struct {
	ID string `json:"id"`
}

func NewAdvertiser(config basil.Config) *Advertiser {
	return &Advertiser{config: config}
}

func (a *Advertiser) Update(state *State) {
	a.Lock()
	defer a.Unlock()

	a.state = state
}

func (a *Advertiser) AdvertisePeriodically(mbus cfmessagebus.MessageBus) {
	interval := time.Duration(a.config.AdvertiseInterval) * time.Second

	go func() {
		for {
			select {
			case <-time.After(interval):
				a.sendAdvertisement(mbus)
			}
		}
	}()
}

func (a *Advertiser) sendAdvertisement(mbus cfmessagebus.MessageBus) {
	a.RLock()
	defer a.RUnlock()

	if a.state == nil {
		return
	}

	msg, err := json.Marshal(&AdvertiseMessage{ID: a.state.ID})
	if err != nil {
		return
	}

	mbus.Publish("ssh.advertise", msg)
}