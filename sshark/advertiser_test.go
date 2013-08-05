package basil_sshark

import (
	"github.com/cloudfoundry/go_cfmessagebus/mock_cfmessagebus"
	"github.com/vito/basil"
	. "launchpad.net/gocheck"
	"time"
)

type ASuite struct{}

func init() {
	Suite(&ASuite{})
}

func (a *ASuite) TestAdvertiserAdvertisesID(c *C) {
	mbus := mock_cfmessagebus.NewMockMessageBus()

	config := basil.Config{
		AdvertiseInterval: 100 * time.Millisecond,
	}

	advertisements := make(chan []byte)

	mbus.Subscribe("ssh.advertise", func(msg []byte) {
		advertisements <- msg
	})

	advertiser := NewAdvertiser(config)

	advertiser.Update(&State{ID: "some-unique-id"})

	advertiser.AdvertisePeriodically(mbus)

	ad := waitReceive(advertisements, 1*time.Second)
	c.Assert(string(ad), Equals, `{"id":"some-unique-id"}`)
}

func (a *ASuite) TestAdvertiserAdvertisesPeriodically(c *C) {
	mbus := mock_cfmessagebus.NewMockMessageBus()

	config := basil.Config{
		AdvertiseInterval: 100 * time.Millisecond,
	}

	advertisements := make(chan []byte)

	mbus.Subscribe("ssh.advertise", func(msg []byte) {
		advertisements <- msg
	})

	advertiser := NewAdvertiser(config)

	advertiser.Update(&State{ID: "some-unique-id"})

	advertiser.AdvertisePeriodically(mbus)

	msg1 := waitReceive(advertisements, 1*time.Second)
	c.Assert(msg1, NotNil)

	time1 := time.Now()

	msg2 := waitReceive(advertisements, 1*time.Second)
	c.Assert(msg2, NotNil)

	time2 := time.Now()

	c.Assert(time2.Sub(time1) >= 100*time.Millisecond, Equals, true)
}

func (a *ASuite) TestAdvertiserAdvertisesUpdatedID(c *C) {
	mbus := mock_cfmessagebus.NewMockMessageBus()

	config := basil.Config{
		AdvertiseInterval: 100 * time.Millisecond,
	}

	advertisements := make(chan []byte)

	mbus.Subscribe("ssh.advertise", func(msg []byte) {
		advertisements <- msg
	})

	advertiser := NewAdvertiser(config)

	advertiser.Update(&State{ID: "some-unique-id"})

	advertiser.AdvertisePeriodically(mbus)

	msg1 := waitReceive(advertisements, 1*time.Second)
	c.Assert(string(msg1), Equals, `{"id":"some-unique-id"}`)

	advertiser.Update(&State{ID: "some-other-unique-id"})

	msg2 := waitReceive(advertisements, 1*time.Second)
	c.Assert(string(msg2), Equals, `{"id":"some-other-unique-id"}`)
}

func (a *ASuite) TestAdvertiserDoesNotAdvertiseWithoutState(c *C) {
	mbus := mock_cfmessagebus.NewMockMessageBus()

	config := basil.Config{
		AdvertiseInterval: 100 * time.Millisecond,
	}

	advertisements := make(chan []byte)

	mbus.Subscribe("ssh.advertise", func(msg []byte) {
		advertisements <- msg
	})

	advertiser := NewAdvertiser(config)
	advertiser.AdvertisePeriodically(mbus)

	msg1 := waitReceive(advertisements, 1*time.Second)
	c.Assert(msg1, IsNil)
}
