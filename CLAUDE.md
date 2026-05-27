# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Build for current platform
make compile

# Cross-compile for all targets (linux amd64, darwin amd64/arm64)
make commit

# Format code (installs goimports if missing)
make format

# Lint (installs staticcheck if missing)
make lint

# Run tests with coverage report
make test

# Run a single test
go test -run TestName ./...

# Format + lint + test + compile in one shot
make precommit

# Update dependencies
go mod tidy
```

## Architecture

The project is a Prometheus exporter for OpenLDAP. It lives in module `github.com/majimenez-stratio/openldap_exporter` and has three Go files at the root package plus the `cmd/` entrypoint:

- **`scraper.go`** — core scraping logic. Defines `Scraper` and `LDAPConfig`, the Prometheus gauges/counters, and the `query` struct. On each tick, `Scraper.scrape()` dials LDAP via `ldap.DialURL`, optionally binds, runs all `queries` against `cn=Monitor`, and updates the gauges. Replication queries are added dynamically via `addReplicationQueries()` based on the `--replicationObject` flag.
- **`server.go`** — HTTP server wrapping `prometheus/exporter-toolkit`. Exposes `/metrics` (via `promhttp`) and `/version`. Uses a `logrusHandler` bridge to adapt `slog` (required by exporter-toolkit) to logrus.
- **`cmd/openldap_exporter/main.go`** — CLI entrypoint using `urfave/cli/v2` with `altsrc` for layered config (flags > env vars > YAML file > defaults). Wires `Server` and `Scraper` together, runs them concurrently via `errgroup`, and handles `SIGINT`/`SIGTERM` for graceful shutdown.

### Configuration precedence

CLI flags → environment variables → YAML file (`--config`) → defaults.

### Key metrics scraped

| Prometheus metric | LDAP objectClass | Attribute |
|---|---|---|
| `openldap_monitored_object` | `monitoredObject` | `monitoredInfo` |
| `openldap_monitor_counter_object` | `monitorCounterObject` | `monitorCounter` |
| `openldap_monitor_operation` | `monitorOperation` | `monitorOpCompleted` |
| `openldap_monitor_replication` | `*` (on each `--replicationObject` DN) | `contextCSN` |

Internal counters (`openldap_dial`, `openldap_bind`, `openldap_scrape`) track connection health.

### TLS handling

`LDAPConfig.ProcessTLSoptions` resolves the scheme from the `--ldapAddr` URI (`ldap://`, `ldaps://`, `ldapi://`) and sets default ports (389/636). `ldaps://` → full TLS with `ldap.DialWithTLSConfig`; `ldap://` + `--ldapUseStartTLS` → StartTLS upgrade after dial.

### LDAP client library

Uses `github.com/go-ldap/ldap/v3` (migrated from the abandoned `gopkg.in/ldap.v2`). Connections are opened with `ldap.DialURL(scheme://addr)`.

### Version injection

`tag` and `commit` in `server.go` are set at link time via `-ldflags` in the Makefile.

### Tests

Unit tests cover `LDAPConfig.ProcessTLSoptions` (all schemes and TLS modes), `LDAPConfig.LoadCACert`, and the `/version` HTTP endpoint. The `make test` target excludes `cmd/` (no test files there) and writes `cover.out`.
