# Development notes

## pprof (optional, dev-only)

This project includes an opt-in pprof server that is disabled by default.

To enable pprof locally set the environment variable `ENABLE_PPROF=1` and run the service. The pprof server will bind to `localhost:6060` (only) to avoid exposing profiling endpoints publicly.

Example (local):

```bash
export ENABLE_PPROF=1
go run ./cmd
# or when running the built binary
./bin/wallet-api

# then open profiling endpoints in your browser or use go tool pprof
# http://localhost:6060/debug/pprof/
# go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

Security note: never enable pprof on a public interface in production without proper access controls and authentication.
