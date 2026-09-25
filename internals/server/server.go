package server

import (
	"io"
	"net/http"
	"net/url"
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
		return
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
	if isRedirectStatus(response.Status) {
		handleRedirection(&response.Headers, appName, s.Tunnel.Domain)
	}
	agent.CopyHeaders(w.Header(), response.Headers)
	w.WriteHeader(response.Status)
	w.Write(response.Body)
}

func isRedirectStatus(status int) bool {
	switch status {
	case http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
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

func handleRedirection(headers *map[string][]string, appName string, domain string) {
	locations := (*headers)["Location"]
	if len(locations) == 0 {
		return
	}

	redirectURL, err := url.Parse(locations[0])
	if err != nil || !redirectURL.IsAbs() {
		return
	}

	redirectURL.Scheme = "http"
	redirectURL.Host = appName + "." + domain
	locations[0] = redirectURL.String()
}
