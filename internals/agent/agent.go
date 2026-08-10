package agent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"sgrok/internals/proto"
	"sgrok/internals/types"
)

type Agent struct {
	TunnelAddr string
	Conn       net.Conn
	ClientID   string
	Port       string
	Mu         *sync.RWMutex
}

func (a Agent) Init() (net.Conn, error) {
	conn, err := net.Dial("tcp", a.TunnelAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (a *Agent) Listen() {
	http := Httpx{
		Port: a.Port,
	}

	http.Init()
	go func() {
		for {
			a.Mu.RLock()
			buffer, err := proto.ReadFrame(a.Conn)
			a.Mu.RUnlock()
			if err != nil {
				fmt.Println(err)
				a.Conn.Close()
				break
			}

			go func() {
				var tunnelRequest types.TunnelRequest
				err = json.Unmarshal(buffer, &tunnelRequest)
				if err != nil {
					fmt.Println(err)
					a.sendResponse(a.Conn, []byte("something went wrong"))
					return
				}

				slog.Info(tunnelRequest.Method, ":", tunnelRequest.Path)
				response, err := http.Request(tunnelRequest)
				response.RequestID = tunnelRequest.ID
				if err != nil {
					a.Conn.Write([]byte("something went wrong"))
					return
				}

				paylaod, err := json.Marshal(response)
				if err != nil {
					a.Conn.Write([]byte("something went wrong"))
					return
				}
				if err = a.sendResponse(a.Conn, paylaod); err != nil {
					fmt.Println(err)
					a.Conn.Close()
				}
			}()

		}
	}()
}

func (a *Agent) sendResponse(conn net.Conn, data []byte) error {
	err := proto.WriteFrame(conn, data)
	return err
}

func (a Agent) Ping(ClientID string) (string, error) {
	request, err := json.Marshal(types.ClientPingRequest{
		Ping:     "ping",
		ClientID: ClientID,
	})
	if err != nil {
		a.Conn.Close()
		return "", err
	}
	err = proto.WriteFrame(a.Conn, request)
	if err != nil {
		return "", err
	}
	buffer, err := proto.ReadFrame(a.Conn)
	if err != nil {
		return "", err
	}
	var payload types.ClientInitRespose

	err = json.Unmarshal(buffer, &payload)
	if err != nil {
		return "", err
	}

	return payload.ClientID, nil
}

func (a Agent) Register(name string) (string, error) {
	request := types.ClientAppNameRequest{
		ClientID: a.ClientID,
		AppName:  name,
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	err = proto.WriteFrame(a.Conn, jsonRequest)
	if err != nil {
		return "", err
	}

	var response types.CientAppNameResponse

	data, err := proto.ReadFrame(a.Conn)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return "", err
	}

	return response.AppURL, nil
}
