package agent

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"sgrok/internals/types"
)

type Httpx struct {
	Port   string
	Client http.Client
}

func (h *Httpx) Init() {
	h.Client = http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (h *Httpx) Request(tunnelRequest types.TunnelRequest) (types.TunnelResponse, error) {
	CopyHeaders(http.Header{}, tunnelRequest.Headers)
	req, err := http.NewRequest(tunnelRequest.Method, "http://localhost:"+string(h.Port)+tunnelRequest.Path, bytes.NewReader(tunnelRequest.Body))
	if err != nil {
		return types.TunnelResponse{}, err
	}
	req.Header = tunnelRequest.Headers
	resp, err := h.Client.Do(req)
	if err != nil {
		return types.TunnelResponse{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.TunnelResponse{}, err
	}

	tunnelResponse := types.TunnelResponse{
		Headers: resp.Header,
		Body:    body,
		Status:  resp.StatusCode,
	}
	return tunnelResponse, nil
}

func CopyHeaders(dist http.Header, src map[string][]string) {
	for key, values := range src {
		for _, value := range values {
			dist.Add(key, value)
		}
	}
}
