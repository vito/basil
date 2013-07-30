package basil

import (
  "sync"
)

type Usage struct {
  MemoryInBytes int
  DiskInBytes int
}

type ResourcePool struct {
  config Config

  consumers []Consumer

  sync.RWMutex
}

type Consumer interface {
  CurrentUsage() Usage
}

func NewResourcePool(config Config) *ResourcePool {
  return &ResourcePool{config: config}
}

func (a *ResourcePool) AddConsumer(consumer Consumer) {
  a.Lock()
  defer a.Unlock()

  a.consumers = append(a.consumers, consumer)
}

func (a *ResourcePool) AvailableMemory() int {
  a.RLock()
  defer a.RUnlock()

  capacity := a.config.Capacity.MemoryInBytes

  for _, consumer := range a.consumers {
    capacity -= consumer.CurrentUsage().MemoryInBytes
  }

  return capacity
}

func (a *ResourcePool) AvailableDisk() int {
  a.RLock()
  defer a.RUnlock()

  capacity := a.config.Capacity.DiskInBytes

  for _, consumer := range a.consumers {
    capacity -= consumer.CurrentUsage().DiskInBytes
  }

  return capacity
}