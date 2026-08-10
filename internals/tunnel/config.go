package tunnel

import (
	"encoding/json"
	"os"
)

type Config struct {
	TunnelPort int    `json:"tunnel_port"`
	Domain     string `json:"domain"`
	ServerPort int    `json:"server_port"`
}

func (config *Config) LoadConfig() error {
	data, err := os.ReadFile("./config.json")
	if err != nil {
		return err
	}

	if data != nil {
		err = json.Unmarshal(data, &config)
	}
	return nil
}
