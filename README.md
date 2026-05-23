<p align="center">
  <h1 align="center">SentinelMesh</h1>
  <p align="center">
    <strong>AI-Powered Threat Intelligence & Attack Surface Analysis Platform</strong>
  </p>
  <p align="center">
    Give it a domain, IP, or CIDR range. Five autonomous agents investigate in parallel, correlate findings, and produce actionable security reports with MITRE ATT&CK mappings and detection rules.
  </p>
</p>

<p align="center">
  <a href="https://github.com/ahmedimran35/SentinelMesh/actions"><img src="https://github.com/ahmedimran35/SentinelMesh/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/ahmedimran35/SentinelMesh/releases"><img src="https://img.shields.io/github/v/release/ahmedimran35/SentinelMesh" alt="Release"></a>
  <img src="https://img.shields.io/badge/language-Go-00ADD8.svg" alt="Go">
  <img src="https://img.shields.io/badge/frontend-React-61DAFB.svg" alt="React">
</p>

---

## How It Works

```
                    Target (domain / IP / CIDR)
                              |
              +-------+-------+-------+-------+
              |       |       |       |       |
           Recon    Vuln   Malware  Threat   News
           Agent    Agent    Agent   Intel    Intel
              |       |       |       |       |
              +-------+-------+-------+-------+
                              |
                     Correlation Engine
                              |
                    +---------+---------+
                    |         |         |
               Report    Alerts    Sigma/YARA
                                    Detection Rules
```

Five specialized agents run concurrently:

| Agent | What It Does | Data Sources |
|-------|-------------|--------------|
| **Recon** | DNS records, subdomains, ports, geo, ASN | Cloudflare DNS, crt.sh, Shodan InternetDB, IPWhois, BGPView |
| **Vuln** | CVE search, severity scoring, exploit tracking | NVD API + LLM analysis |
| **Malware** | IOC matching, malware samples, C2 indicators | ThreatFox, URLHaus, MalwareBazaar |
| **Threat Intel** | LLM-powered threat assessment, MITRE ATT&CK mapping | NVIDIA NIM / Ollama |
| **News Intel** | Exploit PoCs, breach news, zero-day tracking | HackerNews, GitHub, RSS, Reddit |

## Quick Start

### Docker (recommended)

```bash
git clone https://github.com/ahmedimran35/SentinelMesh.git
cd SentinelMesh

# With Ollama (local LLM, no API key needed)
docker compose up

# With NVIDIA NIM (cloud LLM)
NIM_API_KEY=nvapi-xxx docker compose up
```

Open `http://localhost:8090`.

### From Source

```bash
# Requires Go 1.22+ and Node.js 20+
git clone https://github.com/ahmedimran35/SentinelMesh.git
cd SentinelMesh
make run
```

## API

```bash
# Start an investigation
curl -X POST http://localhost:8090/api/investigate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-key" \
  -d '{"target": "example.com", "type": "domain"}'

# List investigations
curl http://localhost:8090/api/investigations

# Get findings
curl http://localhost:8090/api/investigations/{id}/findings

# Export report (JSON, CSV, STIX)
curl http://localhost:8090/api/investigations/{id}/export?format=json

# Real-time events (SSE)
curl http://localhost:8090/events

# Add scheduled monitor
curl -X POST http://localhost:8090/api/monitors \
  -H "Content-Type: application/json" \
  -d '{"target": "example.com", "type": "domain", "scan_interval": "12h"}'
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `API_KEY` | _(empty)_ | API key for auth (**required in production**) |
| `LLM_PROVIDER` | `nim` | LLM backend: `nim` or `ollama` |
| `NIM_API_KEY` | _(empty)_ | NVIDIA NIM API key |
| `NIM_MODEL` | `meta/llama-3.1-70b-instruct` | NIM model |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama URL |
| `OLLAMA_MODEL` | `llama3.1` | Ollama model |
| `PORT` | `8090` | Server port |
| `DB_PATH` | `./sentinelmesh.db` | SQLite database path |
| `API_RATE_LIMIT` | `10` | Requests per minute per IP |
| `MAX_CONCURRENT_SCANS` | `5` | Max parallel investigations |
| `CORS_ORIGINS` | `http://localhost:5173,...` | Allowed CORS origins |
| `TLS_CERT_FILE` | _(empty)_ | TLS cert (enables HTTPS) |
| `TLS_KEY_FILE` | _(empty)_ | TLS key |

## Tech Stack

- **Backend**: Go 1.22 (net/http, SQLite, goroutines)
- **Frontend**: React 18 + Vite (embedded in Go binary via `embed.FS`)
- **Database**: SQLite (WAL mode, zero-config)
- **LLM**: NVIDIA NIM (Llama 3.1 70B) or Ollama (local, free)
- **All data sources**: Free, no API keys required

## Security

- API key auth (header only, no query string leak)
- Configurable CORS allowlist
- Per-IP rate limiting with memory-safe cleanup
- Input validation on all endpoints
- 10MB response body limits
- Crypto-random IDs
- Prompt injection sanitization
- TLS 1.2+ support

## Project Structure

```
.
├── main.go                 # Entry point, routes, server
├── agents/                 # Autonomous investigation agents
│   ├── framework.go        # Agent interface + parallel runner
│   ├── recon.go            # Reconnaissance agent
│   ├── vuln.go             # CVE/vulnerability agent
│   ├── malware.go          # Malware IOC agent
│   ├── threatintel.go      # LLM threat intel agent
│   ├── newsintel.go        # Security news agent
│   └── commander.go        # Report generator + rule engine
├── fetchers/               # OSINT data source clients (14 fetchers)
├── llm/                    # LLM abstraction (NIM + Ollama)
├── correlation/            # Cross-agent pattern detection
├── models/                 # Data models
├── server/                 # HTTP handlers, middleware, SSE
├── store/                  # SQLite storage layer
├── monitor/                # Scheduled scan engine
├── config/                 # Environment config
├── frontend/               # React SPA
├── Dockerfile              # Multi-stage build
├── docker-compose.yml      # One-command setup
└── Makefile                # Build targets
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, workflow, and guidelines.

## License

[MIT](LICENSE)
