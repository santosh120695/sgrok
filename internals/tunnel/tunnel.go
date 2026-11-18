package tunnel

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"

	"sgrok/internals/proto"
	"sgrok/internals/types"

	"github.com/google/uuid"
)

type Tunnel struct {
	Port          int
	Domain        string
	Apps          map[string]types.App
	Mu            *sync.RWMutex
	WriterChannel chan types.WriterChannelType
}

func (t *Tunnel) Init() error {
	listner, err := net.ListenTCP("tcp", &net.TCPAddr{Port: t.Port})
	if err != nil {
		return err
	}
	t.Apps = make(map[string]types.App)
	t.WriterChannel = make(chan types.WriterChannelType, 1000)
	t.Mu = &sync.RWMutex{}
	go t.tunnelWriter()

	go func() {
		for {
			conn, err := listner.Accept()
			if err != nil {
				conn.Close()
				continue
			}
			go t.handleClient(conn)
		}
	}()
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

			err = proto.WriteFrame(conn, response)
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
				t.Mu.RLock()
				if t.Apps[payload.AppName].ClientID != payload.ClientID {
					conn.Close()
					t.Mu.RUnlock()
					break loop
				}
				t.Mu.RUnlock()
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
		t.Mu.RLock()
		app := t.Apps[appName]
		t.Mu.RUnlock()
		data, err := proto.ReadFrame(app.Conn)
		if err != nil {
			if err == io.EOF {
				app.Conn.Close()
				break
			}
		}
		json.Unmarshal(data, &tunnelResponse)
		app.ConnLock.RLock()
		app.AppChan[tunnelResponse.RequestID] <- tunnelResponse
		app.ConnLock.RUnlock()
	}
}

func (t *Tunnel) tunnelWriter() {
	var writeRequest types.WriterChannelType
	for {
		writeRequest = <-t.WriterChannel
		proto.WriteFrame(writeRequest.Conn, writeRequest.Payload)
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
		ConnLock: &sync.RWMutex{},
		AppChan:  make(map[string]chan types.TunnelResponse),
	}
	return appURL, nil
}

func (t *Tunnel) ForwardRequest(tunnelRequest types.TunnelRequest) (types.TunnelResponse, error) {
	t.Mu.RLock()
	app, ok := t.Apps[tunnelRequest.AppName]
	t.Mu.RUnlock()
	if !ok {
		return types.TunnelResponse{}, errors.New("app not available")
	}
	app.ConnLock.Lock()
	app.AppChan[tunnelRequest.ID] = make(chan types.TunnelResponse)
	readChan := app.AppChan[tunnelRequest.ID]
	app.ConnLock.Unlock()
	payload, err := json.Marshal(tunnelRequest)
	if err != nil {
		return types.TunnelResponse{}, errors.New("error in parsing request")
	}
	t.WriterChannel <- types.WriterChannelType{Conn: app.Conn, Payload: payload}
	response := <-readChan
	return response, nil
}
