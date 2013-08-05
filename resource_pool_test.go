package basil

import (
	"github.com/remogatto/prettytest"
	"testing"
)

type RPSuite struct {
	prettytest.Suite
}

func TestResourcePoolRunner(t *testing.T) {
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(RPSuite),
	)
}

type FakeConsumer struct {
	UsedMemoryInBytes int
	UsedDiskInBytes   int
}

func (c *FakeConsumer) CurrentUsage() Usage {
	return Usage{
		MemoryInBytes: c.UsedMemoryInBytes,
		DiskInBytes:   c.UsedDiskInBytes,
	}
}

func (s *RPSuite) TestResourcePoolAvailableMemory() {
	config := Config{
		Capacity: CapacityConfig{
			MemoryInBytes: 1 * gigabyte,
			DiskInBytes:   1 * gigabyte,
		},
	}

	pool := NewResourcePool(config)
	pool.AddConsumer(&FakeConsumer{UsedMemoryInBytes: 42})
	pool.AddConsumer(&FakeConsumer{UsedMemoryInBytes: 128})

	s.Equal(pool.AvailableMemory(), (1*gigabyte)-42-128)
}

func (s *RPSuite) TestResourcePoolAvailableDisk() {
	config := Config{
		Capacity: CapacityConfig{
			MemoryInBytes: 1 * gigabyte,
			DiskInBytes:   1 * gigabyte,
		},
	}

	pool := NewResourcePool(config)
	pool.AddConsumer(&FakeConsumer{UsedDiskInBytes: 42})
	pool.AddConsumer(&FakeConsumer{UsedDiskInBytes: 128})

	s.Equal(pool.AvailableDisk(), (1*gigabyte)-42-128)
}
