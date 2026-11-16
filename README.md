# sgrok

A lightweight, self-hosted reverse proxy and tunneling tool written in Go, inspired by ngrok. `sgrok` allows you to securely expose local web services (like `localhost:3000`) to the internet or your local network through custom subdomains, without configuring router port-forwarding or dynamic DNS.

---

## Architecture Overview

```
                      +---------------------------------------------------+
                      |                   sgrok Gateway                   |
                      |                                                   |
[ External User ] --->| HTTP Gateway (:5050)       TCP Tunnel (:5051)     |
(myapp.localhost)     |   • Subdomain Router         • Frame Protocol     |
                      |   • Redirect Rewriter        • App Registry       |
                      +-------------------+-------------------------------+
                                          ^
                                          | Persistent TCP Tunnel
                                          v
                      +---------------------------------------------------+
                      |                    sgrok Agent                    |
                      |                                                   |
                      |   • TCP Multiplexed Receiver                      |
                      |   • Local HTTP Dispatcher                         |
                      +-------------------+-------------------------------+
                                          |
                                          v Local HTTP Request
                      +---------------------------------------------------+
                      |             Local Web Application                 |
                      |             (e.g., localhost:3000)                |
                      +---------------------------------------------------+
```

---

## Features

- **Subdomain-based Routing**: Automatically routes traffic based on incoming host subdomains (`http://<app-name>.<domain>`).
- **No Port-Forwarding Required**: The local agent establishes an outbound TCP connection to the gateway, bypassing NATs and firewalls.
- **Header & Redirect Preservation**: Transparently proxies HTTP methods, custom headers, query parameters, status codes, and automatically rewrites `Location` redirect headers.
- **Client Session Persistence**: Automatically remembers and persists `client_id` in `/tmp/sgrok/sgrok.json` to retain app ownership across reconnects.
- **CORS Enabled**: Out-of-the-box cross-origin resource sharing support for API development and webhooks.
- **Load Test Ready**: Includes Grafana k6 performance testing scripts to benchmark throughput.

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.22+ installed
- A local web application running (e.g. on port `3000`)

### Installation & Build

Clone the repository and compile the binaries:

```bash
git clone https://github.com/santoshghode/sgrok.git
cd sgrok

# Build the gateway server
go build -o sgrok_server cmd/gateway/main.go

# Build the agent CLI
go build -o sgrok cmd/agent/main.go
```

Alternatively, use the build script:

```bash
chmod +x build.sh
./build.sh
```

---

## Configuration

### 1. Gateway Server Configuration (`config.json`)

Place `config.json` in the root directory where `sgrok_server` runs:

```json
{
  "server_port": 5050,
  "domain": "localhost:5050",
  "tunnel_port": 5051
}
```

- `server_port`: The HTTP port where the public/external requests arrive.
- `domain`: Base domain for host matching and subdomain routing.
- `tunnel_port`: The internal TCP port the agent connects to.

### 2. Agent Configuration (`agent.json`)

Place `agent.json` in the root directory where `sgrok` runs:

```json
{
  "tunnel_address": "localhost:5051"
}
```

- `tunnel_address`: The `host:port` of the running gateway tunnel server.

---

## Usage Guide

### Step 1: Start your Local Web Application

Ensure your local backend or frontend server is running:

```bash
# Example: Running a simple HTTP server on port 3000
python3 -m http.server 3000
```

### Step 2: Start the sgrok Gateway

Run the gateway server on your public server (or local machine):

```bash
./sgrok_server
```

The gateway listens for external HTTP traffic on port `5050` and agent tunnel connections on port `5051`.

### Step 3: Start the sgrok Agent

On your development machine, expose your local service by running:

```bash
./sgrok serve <app_name> <local_port>
```

**Example:** Expose port `3000` under the subdomain `ecom`:

```bash
./sgrok serve ecom 3000
```

**Output:**
```text
>>>>>> Welcome to sgrok <<<<<<<
app url:  http://ecom.localhost:5050
```

### Step 4: Access your Service

Open your browser or make an HTTP request to the provided URL:

```bash
curl http://ecom.localhost:5050/
```

Traffic will flow through the gateway, travel through the TCP tunnel, hit your local server at `localhost:3000`, and return the response.

---

## Project Structure

```text
.
├── cmd/
│   ├── agent/                 # Agent CLI entrypoint
│   └── gateway/               # Gateway server entrypoint
├── internals/
│   ├── agent/                 # Agent state, HTTP dispatcher, and configuration
│   ├── proto/                 # Big-endian length-prefixed TCP frame protocol
│   ├── server/                # Gateway HTTP handler, subdomain resolver, and CORS
│   ├── tunnel/                # Gateway TCP listener and request multiplexer
│   └── types/                 # Shared request/response data models
├── agent.json                 # Agent configuration
├── config.json                # Gateway configuration
├── script.js                  # Grafana k6 load testing suite
└── build.sh                   # Build script
```

---

## License

MIT
