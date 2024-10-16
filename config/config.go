package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Nodes        []Node        `yaml:"nodes"`
	SyncInterval time.Duration `yaml:"sync_interval"`
}

func New(filepath string) Config {
	var config Config

	bs, err := os.ReadFile(filepath)
	if err != nil {
		panic(fmt.Sprintf("failed to read config file: %v", err))
	}
	err = yaml.Unmarshal(bs, &config)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal config content: %v", err))
	}

	config.SyncInterval = time.Duration(config.SyncInterval) * time.Second

	return config
}

func (c *Config) SprintNodeNames() string {
	names := make([]string, 0, len(c.Nodes))
	for _, node := range c.Nodes {
		names = append(names, node.Name)
	}

	return strings.Join(names, "\n")
}

type Node struct {
	ID          int    `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	RPCHost     string `yaml:"rpc_host"`
	RPCPort     int    `yaml:"rpc_port"`
	UseRest     bool   `yaml:"use_rest"`
	RPCUser     string `yaml:"rpc_user"`
	RPCPassword string `yaml:"rpc_password"`
}
