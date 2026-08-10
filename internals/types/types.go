package types

import (
	"net"
	"sync"
)

type ClientInitRespose struct {
	ClientID string `json:"client_id"`
}

type ClientAppNameRequest struct {
	ClientID string `json:"client_id"`
	AppName  string `json:"app_name"`
}

type CientAppNameResponse struct {
	ClientID string `json:"client_id"`
	AppURL   string `json:"app_url"`
}

type ClientPingRequest struct {
	Ping     string `json:"ping"`
	ClientID string `json:"client_id"`
}

type App struct {
	Name     string   `json:"name"`
	ClientID string   `json:"client_id"`
	Conn     net.Conn `json:"conn"`
	ConnLock *sync.Mutex
	AppURL   string                         `json:"app_url"`
	AppChan  map[string]chan TunnelResponse `json:"app_chan"`
}

type TunnelRequest struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	ID      string              `json:"id"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
	AppName string              `json:"app_name"`
}

type TunnelResponse struct {
	Body      []byte              `json:"body"`
	Headers   map[string][]string `json:"headers"`
	Status    string              `json:"status"`
	RequestID string              `json:"request_id"`
}
