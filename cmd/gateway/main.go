package main

import (
	"fmt"

	"sgrok/internals/server"
	"sgrok/internals/tunnel"
)

func main() {
	config := tunnel.Config{}
	err := config.LoadConfig()
	if err != nil {
		fmt.Println(err)
	}

	tunnelServer := tunnel.Tunnel{
		Port:   config.TunnelPort,
		Domain: config.Domain,
	}

	gatewayServer := server.Server{
		Port:   config.ServerPort,
		Host:   "localhost",
		Tunnel: &tunnelServer,
	}

	tunnelServer.Init()

	if err := gatewayServer.Init(); err != nil {
		fmt.Println(err)
	}
}
