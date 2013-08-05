package basil_sshark

import (
	"time"
)

func timedReceive(from chan time.Time, giveup time.Duration) *time.Time {
	select {
	case val := <-from:
		return &val
	case <-time.After(giveup):
		return nil
	}
}

func waitReceive(from chan []byte, giveup time.Duration) []byte {
	select {
	case val := <-from:
		return val
	case <-time.After(giveup):
		return nil
	}
}
