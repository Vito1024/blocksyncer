package config

import (
	"testing"
)

func TestNew(t *testing.T) {
	config := New("./config.yaml")
	t.Logf("config: %v\n", config)
	t.Logf("node names: \n%v\n", config.SprintNodeNames())
}
