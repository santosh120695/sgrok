package main

import (
	"fmt"
	"os"
	"sync"

	"sgrok/internals/agent"

	"github.com/spf13/cobra"
)

func main() {
	fmt.Println(">>>>>> Welcome to sgrok <<<<<<<")
	var wg sync.WaitGroup

	initCmd := &cobra.Command{
		Use:   "serve",
		Short: "start serving app ",
		Run: func(cmd *cobra.Command, args []string) {
			config := agent.Config{}
			config.Load()
			a := agent.Agent{
				Mu:         &sync.RWMutex{},
				TunnelAddr: config.TunnelAddress,
			}
			state := agent.State{}
			state.Load()

			conn, err := a.Init()
			if err != nil {
				fmt.Println("Error in connecting with tunnel")
				return
			}
			a.Conn = conn

			clientID, err := a.Ping(state.ClientID)
			if err != nil {
				fmt.Println(err)
				return
			}

			a.ClientID = clientID
			AppURL, err := a.Register(args[0])
			a.Port = args[1]
			if err != nil {
				fmt.Println(err)
			}
			state.ClientID = clientID
			state.Save()
			fmt.Println("app url: ", AppURL)
			wg.Add(1)
			a.Listen()
			wg.Wait()
		},
	}

	rootCmd := &cobra.Command{
		Use:   "sgrok agent",
		Short: "use for registring app in tunnel",
	}

	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
