package server

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"sgrok/internals/agent"
	"sgrok/internals/tunnel"
	"sgrok/internals/types"

	"github.com/google/uuid"
	"github.com/rs/cors"
)

type Server struct {
	Port   int
	Host   string
	Tunnel *tunnel.Tunnel
}

func (s *Server) Init() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	c := cors.AllowAll()
	handler := c.Handler(mux)
	err := http.ListenAndServe(strings.Join([]string{s.Host, strconv.Itoa(s.Port)}, ":"), handler)

	return err
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	appName, err := appNameFromDomain(r.Host)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("error in reading body"))
	}

	tunnelRequest := types.TunnelRequest{
		Body:    body,
		Path:    r.URL.RequestURI(),
		Headers: r.Header,
		Method:  r.Method,
		AppName: appName,
		ID:      uuid.New().String(),
	}

	response, err := s.Tunnel.ForwardRequest(tunnelRequest)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	agent.CopyHeaders(w.Header(), response.Headers)
	w.Write([]byte(response.Body))
}

func appNameFromDomain(host string) (string, error) {
	host = strings.TrimPrefix(strings.TrimPrefix(host, "http://"), "https://")

	domains := strings.Split(host, ".")

	if len(domains) <= 2 {
		return domains[0], nil
	}

	if len(domains) == 3 {
		return domains[0], nil
	}

	if len(domains) > 3 {
		return domains[len(domains)-3], nil
	}
	return "", nil
}
