package basil

import (
  . "launchpad.net/gocheck"
)

type RPSuite struct{}

func init() {
  Suite(&RPSuite{})
}

type FakeConsumer struct {
  UsedMemoryInBytes int
  UsedDiskInBytes int
}

func (c *FakeConsumer) CurrentUsage() Usage {
  return Usage{
    MemoryInBytes: c.UsedMemoryInBytes,
    DiskInBytes: c.UsedDiskInBytes,
  }
}

func (s *RPSuite) TestResourcePoolAvailableMemory(c *C) {
  config := Config{
    Capacity: CapacityConfig{
      MemoryInBytes: 1 * gigabyte,
      DiskInBytes:   1 * gigabyte,
    },
  }

  pool := NewResourcePool(config)
  pool.AddConsumer(&FakeConsumer{UsedMemoryInBytes: 42})
  pool.AddConsumer(&FakeConsumer{UsedMemoryInBytes: 128})

  c.Assert(pool.AvailableMemory(), Equals, (1 * gigabyte) - 42 - 128)
}

func (s *RPSuite) TestResourcePoolAvailableDisk(c *C) {
  config := Config{
    Capacity: CapacityConfig{
      MemoryInBytes: 1 * gigabyte,
      DiskInBytes:   1 * gigabyte,
    },
  }

  pool := NewResourcePool(config)
  pool.AddConsumer(&FakeConsumer{UsedDiskInBytes: 42})
  pool.AddConsumer(&FakeConsumer{UsedDiskInBytes: 128})

  c.Assert(pool.AvailableDisk(), Equals, (1 * gigabyte) - 42 - 128)
}