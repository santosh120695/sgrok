package tunnel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"sgrok/internals/proto"
	"sgrok/internals/types"

	"github.com/google/uuid"
)

type Tunnel struct {
	Port   int
	Domain string
	Apps   map[string]types.App
	Mu     *sync.RWMutex
}

func (t *Tunnel) Init() error {
	listner, err := net.ListenTCP("tcp", &net.TCPAddr{Port: t.Port})
	if err != nil {
		return err
	}
	t.Apps = make(map[string]types.App)
	t.Mu = &sync.RWMutex{}

	go func() {
		for {
			conn, err := listner.Accept()
			if err != nil {
				conn.Close()
			}
			go t.handleClient(conn)
		}
	}()
	fmt.Println("starting gateway -> ", t.Port)
	return nil
}

func (t *Tunnel) handleClient(conn net.Conn) {
	state := "INIT"
loop:
	for {
		buffer, err := proto.ReadFrame(conn)
		if err != nil {
			conn.Close()
			break loop
		}

		switch state {
		case "INIT":
			var clientPingRequest types.ClientPingRequest

			err := json.Unmarshal(buffer, &clientPingRequest)
			if err != nil {
				conn.Close()
				break loop
			}

			if clientPingRequest.ClientID == "" {
				clientPingRequest.ClientID = generateClientID().String()
			}
			response, err := json.Marshal(types.ClientInitRespose{
				ClientID: clientPingRequest.ClientID,
			})
			if err != nil {
				conn.Close()
				break loop
			}

			err = proto.WriteFrame(conn, []byte(response))
			if err != nil {
				conn.Close()
				break loop
			}
			state = "REGISTER"

		case "REGISTER":
			var payload types.ClientAppNameRequest
			err := json.Unmarshal(buffer, &payload)
			if err != nil {
				conn.Close()
				break loop
			}
			if t.appNameAvailable(payload.AppName) {
				// todo: send error response to client
				conn.Close()
				break loop
			}
			appURL, err := t.registerNewApp(payload.AppName, payload.ClientID, conn)
			response := types.CientAppNameResponse{
				ClientID: payload.ClientID,
				AppURL:   appURL,
			}

			if err != nil {
				conn.Close()
				break loop
			}

			jsonResponse, _ := json.Marshal(response)
			err = proto.WriteFrame(conn, jsonResponse)
			if err != nil {
				conn.Close()
				break loop
			}
			go t.tunnelReader(payload.AppName)
			break loop

		case "END":
			break loop
		default:
			break loop

		}
	}
}

func generateClientID() uuid.UUID {
	return uuid.New()
}

func (t *Tunnel) tunnelReader(appName string) {
	for {
		var tunnelResponse types.TunnelResponse
		app := t.Apps[appName]
		data, err := proto.ReadFrame(t.Apps[appName].Conn)
		if err != nil {
			if err == io.EOF {
				t.Apps[appName].Conn.Close()
				break
			}
		}

		json.Unmarshal(data, &tunnelResponse)
		app.ConnLock.Lock()
		app.AppChan[tunnelResponse.RequestID] <- tunnelResponse
		app.ConnLock.Unlock()

	}
}

func (t *Tunnel) appNameAvailable(name string) bool {
	t.Mu.RLock()
	defer t.Mu.RUnlock()
	return t.Apps[name].Name != ""
}

func (t *Tunnel) registerNewApp(name string, clientID string, conn net.Conn) (string, error) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	appURL := "http://" + name + "." + t.Domain
	t.Apps[name] = types.App{
		Name:     name,
		ClientID: clientID,
		Conn:     conn,
		AppURL:   appURL,
		ConnLock: &sync.Mutex{},
		AppChan:  make(map[string]chan types.TunnelResponse),
	}
	return appURL, nil
}

func (t *Tunnel) ForwardRequest(tunnelRequest types.TunnelRequest) (types.TunnelResponse, error) {
	t.Mu.Lock()
	app, ok := t.Apps[tunnelRequest.AppName]
	if !ok {
		return types.TunnelResponse{}, errors.New("app not available")
	}
	app.ConnLock.Lock()
	app.AppChan[tunnelRequest.ID] = make(chan types.TunnelResponse)
	app.ConnLock.Unlock()
	t.Mu.Unlock()
	payload, err := json.Marshal(tunnelRequest)
	if err != nil {
		// todo: handle error
		return types.TunnelResponse{}, errors.New("error in parsing request")
	}

	err = proto.WriteFrame(app.Conn, payload)
	if err != nil {
		return types.TunnelResponse{}, errors.New("error in server connection")
	}
	// app.ConnLock.Lock()
	response := <-app.AppChan[tunnelRequest.ID]
	// app.ConnLock.Unlock()
	return response, nil
}
