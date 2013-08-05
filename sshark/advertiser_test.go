package basil_sshark

import (
	"github.com/cloudfoundry/go_cfmessagebus/mock_cfmessagebus"
	"github.com/remogatto/prettytest"
	"github.com/vito/basil"
	"testing"
	"time"
)

type ASuite struct {
	prettytest.Suite
}

func TestAdvertiserRunner(t *testing.T) {
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(ASuite),
	)
}

func (s *ASuite) TestAdvertiserAdvertisesID() {
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
	s.Equal(string(ad), `{"id":"some-unique-id"}`)
}

func (s *ASuite) TestAdvertiserAdvertisesPeriodically() {
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
	s.Not(s.Nil(msg1))

	time1 := time.Now()

	msg2 := waitReceive(advertisements, 1*time.Second)
	s.Not(s.Nil(msg2))

	time2 := time.Now()

	s.True(time2.Sub(time1) >= 100*time.Millisecond)
}

func (s *ASuite) TestAdvertiserAdvertisesUpdatedID() {
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
	s.Equal(string(msg1), `{"id":"some-unique-id"}`)

	advertiser.Update(&State{ID: "some-other-unique-id"})

	msg2 := waitReceive(advertisements, 1*time.Second)
	s.Equal(string(msg2), `{"id":"some-other-unique-id"}`)
}

func (s *ASuite) TestAdvertiserDoesNotAdvertiseWithoutState() {
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
	s.Nil(msg1)
}
