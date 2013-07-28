package basil

import (
	"github.com/kylelemons/go-gypsy/yaml"
	"strconv"
)

type Config struct {
	Host       string
	MessageBus MessageBusConfig
	Capacity   CapacityConfig
}

type MessageBusConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type CapacityConfig struct {
	MemoryInBytes int
	DiskInBytes   int
}

var kilobyte = 1024
var megabyte = kilobyte * 1024
var gigabyte = megabyte * 1024

var DefaultConfig = Config{
	Host: "127.0.0.1",

	MessageBus: MessageBusConfig{
		Host: "127.0.0.1",
		Port: 4222,
	},

	Capacity: CapacityConfig{
		MemoryInBytes: 1 * gigabyte,
		DiskInBytes:   1 * gigabyte,
	},
}

func LoadConfig(configFilePath string) Config {
	file := yaml.ConfigFile(configFilePath)

	host := file.Require("host")

	mbusHost := file.Require("message_bus.host")
	mbusPort, err := strconv.Atoi(file.Require("message_bus.port"))
	if err != nil {
		panic("non-numeric message bus port")
	}

	mbusUsername, _ := file.Get("message_bus.username")
	mbusPassword, _ := file.Get("message_bus.password")

	capacityMemory, err := strconv.Atoi(file.Require("capacity.memory"))
	if err != nil {
		panic("non-numeric memory capacity")
	}

	capacityDisk, err := strconv.Atoi(file.Require("capacity.disk"))
	if err != nil {
		panic("non-numeric disk capacity")
	}

	return Config{
		Host: host,

		MessageBus: MessageBusConfig{
			Host:     mbusHost,
			Port:     mbusPort,
			Username: mbusUsername,
			Password: mbusPassword,
		},

		Capacity: CapacityConfig{
			MemoryInBytes: capacityMemory * 1024 * 1024,
			DiskInBytes:   capacityDisk * 1024 * 1024,
		},
	}
}
