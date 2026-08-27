package agent

import (
	"encoding/json"
	"os"
)

type Config struct {
	TunnelAddress string `json:"tunnel_address"`
}

func (config *Config) Load() error {
	data, err := os.ReadFile("./agent.json")
	if data != nil {
		err = json.Unmarshal(data, &config)
	}
	return err
}
