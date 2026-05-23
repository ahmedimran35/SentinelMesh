# SentinelMesh

An autonomous, AI-powered threat intelligence and attack surface analysis platform. Give it a domain, IP, or IP range — it runs multiple specialized agents in parallel, correlates findings, and generates actionable security reports with risk ratings, IOCs, MITRE ATT&CK mappings, and detection rules.

## How It Works

SentinelMesh uses a multi-agent architecture where each agent is an independent specialist. When you submit a target, all agents run concurrently, feeding findings into a correlation engine that detects cross-agent patterns like active exploit campaigns or coordinated attack surfaces.

```
Target (domain/IP/range)
        │
        ├──► Recon Agent        ── DNS, subdomains, ASN, IP geolocation
        ├──► Vuln Agent         ── CVE lookup via NVD, LLM-powered analysis
        ├──► Malware Agent      ── ThreatFox, URLHaus, MalwareBazaar IOC checks
        ├──► Threat Intel Agent ── LLM-driven threat assessment & MITRE mapping
        └──► News Intel Agent   ── HackerNews, GitHub, Reddit, RSS exploit tracking
                │
                ▼
        Correlation Engine ── cross-references all findings
                │
                ▼
        Report + Alerts + Detection Rules
```

## Agents

| Agent | What It Does | Data Sources |
|-------|-------------|--------------|
| **Recon** | Maps attack surface: DNS records, subdomains (crt.sh), IP geolocation, ASN, Shodan-style data | DNS, crt.sh, InternetDB, IPWhois, ASN |
| **Vuln** | Searches for known CVEs, uses LLM to assess exploitability and impact | NVD API + NVIDIA NIM / Ollama |
| **Malware** | Checks IOCs against malware databases, tracks samples and C2 infrastructure | ThreatFox, URLHaus, MalwareBazaar |
| **Threat Intel** | LLM-powered analysis: threat assessment, attack vector prediction, MITRE ATT&CK technique mapping | NVIDIA NIM / Ollama |
| **News Intel** | Monitors security news for exploits, breaches, and zero-days related to the target | HackerNews, GitHub, RSS, Reddit |

## Tech Stack

- **Backend**: Go 1.22 (net/http, SQLite via go-sqlite3)
- **Frontend**: React 18 + Vite (embedded in Go binary via `embed.FS`)
- **Database**: SQLite (WAL mode, zero-config)
- **LLM Providers**: NVIDIA NIM (Llama 3.1 70B) or Ollama (local)
- **Infrastructure**: Docker multi-stage build, GitHub Actions CI

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+ (for frontend build)
- SQLite (included on most systems)

### Run Locally

```bash
# Clone
git clone https://github.com/ahmedimran35/SentinelMesh.git
cd SentinelMesh

# Build and run
make run
```

Server starts at `http://localhost:8090`.

### Run with Docker

```bash
make docker-build
docker run -p 8090:8090 -e LLM_PROVIDER=ollama -e OLLAMA_URL=http://host.docker.internal:11434 sentinelmesh:latest
```

## Configuration

All config via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_PROVIDER` | `nim` | LLM backend: `nim` or `ollama` |
| `NIM_API_KEY` | _(empty)_ | NVIDIA NIM API key |
| `NIM_MODEL` | `meta/llama-3.1-70b-instruct` | NIM model to use |
| `NIM_ENDPOINT` | `https://integrate.api.nvidia.com/v1/chat/completions` | NIM API endpoint |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama server URL |
| `OLLAMA_MODEL` | `llama3.1` | Ollama model to use |
| `PORT` | `8090` | Server port |
| `HOST` | `0.0.0.0` | Bind address |
| `DB_PATH` | `./sentinelmesh.db` | SQLite database path |
| `API_KEY` | _(empty)_ | API key for auth (empty = unauthenticated dev mode) |
| `CORS_ORIGINS` | `http://localhost:5173,http://localhost:3000` | Allowed CORS origins |
| `API_RATE_LIMIT` | `10` | Requests per minute per IP |
| `MAX_CONCURRENT_SCANS` | `5` | Max parallel investigations |
| `DEFAULT_SCAN_INTERVAL` | `24h` | Default monitoring interval |
| `TLS_CERT_FILE` | _(empty)_ | TLS certificate (enables HTTPS) |
| `TLS_KEY_FILE` | _(empty)_ | TLS private key |

## API

### Start Investigation

```bash
curl -X POST http://localhost:8090/api/investigate \
  -H "Content-Type: application/json" \
  -d '{"target": "example.com", "target_type": "domain"}'
```

### List Investigations

```bash
curl http://localhost:8090/api/investigations
```

### Get Investigation Details

```bash
curl http://localhost:8090/api/investigations/{id}
```

### Get Findings

```bash
curl http://localhost:8090/api/investigations/{id}/findings
```

### Search Findings

```bash
curl http://localhost:8090/api/findings/search?q=cve&type=cve&severity=critical
```

### Export Report

```bash
curl http://localhost:8090/api/investigations/{id}/export
```

### Alerts

```bash
# List alerts
curl http://localhost:8090/api/alerts

# Acknowledge alert
curl -X PUT http://localhost:8090/api/alerts/{id}/ack
```

### Monitors (Scheduled Scans)

```bash
# Add monitor
curl -X POST http://localhost:8090/api/monitors \
  -H "Content-Type: application/json" \
  -d '{"target": "example.com", "interval": "12h"}'

# List monitors
curl http://localhost:8090/api/monitors

# Remove monitor
curl -X DELETE http://localhost:8090/api/monitors/{id}
```

### Settings

```bash
# Get settings
curl http://localhost:8090/api/settings

# Save settings
curl -X POST http://localhost:8090/api/settings \
  -H "Content-Type: application/json" \
  -d '{"llm_provider": "ollama", "ollama_model": "llama3.1"}'
```

### Health Check

```bash
curl http://localhost:8090/api/health
```

### Real-time Events (SSE)

```bash
curl http://localhost:8090/events
```

Server-Sent Events stream for live investigation progress updates.

## Security

- **API Key Auth**: Set `API_KEY` env var to protect endpoints. Requests pass via `X-API-Key` header or `?api_key=` query param.
- **CORS**: Configurable origin allowlist via `CORS_ORIGINS`.
- **Rate Limiting**: Per-IP rate limiting (configurable via `API_RATE_LIMIT`).
- **TLS**: Set `TLS_CERT_FILE` and `TLS_KEY_FILE` for HTTPS. Minimum TLS 1.2.
- **Request Limits**: 1MB body size cap on POST endpoints.

## Project Structure

```
.
├── main.go              # Entry point, route registration, server setup
├── agents/              # Autonomous investigation agents
│   ├── framework.go     # Agent interface, registry, parallel runner
│   ├── recon.go         # Reconnaissance agent
│   ├── vuln.go          # Vulnerability/CVE agent
│   ├── malware.go       # Malware IOC agent
│   ├── threatintel.go   # LLM threat intelligence agent
│   └── newsintel.go     # Security news monitoring agent
├── fetchers/            # OSINT data source clients
│   ├── dns.go           # DNS record lookup
│   ├── crtsh.go         # Certificate transparency (subdomain enum)
│   ├── internetdb.go    # Shodan InternetDB
│   ├── ipwhois.go       # IP geolocation/WHOIS
│   ├── asn.go           # ASN lookup
│   ├── nvd.go           # NVD CVE database
│   ├── threatfox.go     # ThreatFox IOC database
│   ├── urlhaus.go       # URLHaus malware URL database
│   ├── malwarebazaar.go # MalwareBazaar sample database
│   ├── hackernews.go    # HackerNews security stories
│   ├── github.go        # GitHub security advisories
│   ├── reddit.go        # Reddit security subreddit
│   └── rss.go           # RSS feed parser
├── llm/                 # LLM provider abstraction
│   ├── provider.go      # Provider interface + factory
│   ├── nim.go           # NVIDIA NIM client
│   └── ollama.go        # Ollama client
├── correlation/         # Cross-agent finding correlation engine
│   └── correlator.go    # Pattern detection across agents
├── models/              # Data models
│   ├── investigation.go # Investigation, Finding, Alert, IOC, Report types
│   └── events.go        # SSE event types
├── server/              # HTTP layer
│   ├── handlers.go      # API endpoint handlers
│   ├── middleware.go     # Auth, CORS, rate limiting
│   └── sse.go           # Server-Sent Events manager
├── store/               # Database layer
│   └── sqlite.go        # SQLite storage with migrations
├── monitor/             # Scheduled scan engine
│   └── scheduler.go     # Cron-like investigation scheduler
├── config/              # Configuration from env vars
│   └── config.go
├── frontend/            # React SPA (embedded in binary)
│   ├── src/
│   ├── package.json
│   └── vite.config.js
├── Dockerfile           # Multi-stage Docker build
├── Makefile             # Build, test, lint, run targets
└── .github/workflows/   # CI pipeline
    └── ci.yml           # Build + test + lint on push/PR
```

## Build Targets

```bash
make build           # Build Go binary (includes frontend)
make test            # Run all Go tests
make lint            # Run golangci-lint
make frontend-build  # Build frontend only
make docker-build    # Build Docker image
make run             # Build and run
make clean           # Remove build artifacts and DB
make install-tools   # Install dev tools (golangci-lint)
```

## Contributing

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/my-change`)
3. Commit with clear messages
4. Push and open a PR

CI runs build, tests, and lint on every push/PR to `main`.

## License

MIT
