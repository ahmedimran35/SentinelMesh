# Contributing to SentinelMesh

Thanks for your interest in contributing! Here's how to get started.

## Quick Setup

```bash
# Clone and enter the repo
git clone https://github.com/ahmedimran35/SentinelMesh.git
cd SentinelMesh

# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend && npm ci && cd ..

# Run tests
go test ./...

# Build and run
make run
```

## Development Workflow

1. Fork the repo
2. Create a branch: `git checkout -b feature/my-change`
3. Make your changes
4. Run tests: `go test -race ./...`
5. Run linter: `golangci-lint run ./...`
6. Commit with a clear message
7. Push and open a PR against `main`

CI will run build, tests, and lint automatically on your PR.

## Code Guidelines

- **Go**: Follow standard Go conventions. Run `golangci-lint` before pushing.
- **Error handling**: Always handle errors. Don't use `_` for error returns except in deferred calls.
- **Concurrency**: Use mutexes or channels properly. Run tests with `-race`.
- **SQL**: Use parameterized queries. Never concatenate user input into SQL.
- **HTTP**: Validate all input at API boundary. Return proper status codes.

## Adding a New Fetcher

1. Create `fetchers/myfetcher.go` implementing your data source client
2. Use `BaseFetcher` for HTTP calls (handles rate limiting and retries)
3. Register it in the appropriate agent or create a new agent
4. Add tests in `fetchers/myfetcher_test.go`

## Adding a New Agent

1. Create `agents/myagent.go` implementing the `Agent` interface:
   ```go
   type Agent interface {
       Name() string
       Description() string
       Investigate(ctx context.Context, target models.Target, findings chan<- models.Finding) error
   }
   ```
2. Register it in `server/handlers.go` `NewHandler()`
3. Add tests

## Reporting Issues

- Use GitHub Issues
- Include steps to reproduce
- Include Go version (`go version`) and OS

## License

By contributing, you agree your contributions will be licensed under the MIT License.
