# AGENTS.md

Freedxm is a Linux port of the Freedom blocker: a Go daemon that blocks
processes and websites for timed sessions, plus a CLI client that talks to it
over gRPC. See `README.md` for user-facing docs.

## Commands

```sh
go build ./...                 # build (binary name: freedxm, gitignored)
go test ./...                  # run tests
revive --formatter stylish ./...   # lint (config: revive.toml)
protoc --go_out=. --go-grpc_out=. ./rpc/server.proto   # regenerate rpc/pb
```

Run `go test ./...` and `revive` before committing; keep lint clean.

## Architecture

- `main.go` — CLI definition (urfave/cli v3): `serve`, `new`, `list`/`ls`,
  `show-config`. All take `--port` (default `8687`).
- `commands/` — CLI actions. `serve.go` is the daemon: a polling loop reads
  `/proc` once per second while any matcher is non-empty, kills matching
  processes, and starts the Chrome monitor when `chrome` is running and
  domains are blocked. The other commands are thin gRPC clients.
- `rpc/` — `server.proto` is the source of truth; `rpc/pb/` is generated
  (never edit by hand). `rpc/implementation/` holds the server and client.
  `CreateSession` looks up block lists in the config, stores the session, and
  schedules its removal with `time.AfterFunc`.
- `model/` — `CurrentSessions` (mutex-protected map) merges all active
  sessions' domains/processes into deduplicated lists.
- `patterns/` — `Matcher`: wildcard patterns where `%` means "anything",
  case-insensitive, unanchored (substring match), guarded by an `RWMutex`.
- `chrome/` — blocks domains via the Chrome DevTools Protocol over a WebSocket
  (gorilla/websocket). Requires Chrome running with remote debugging (port
  `9222`, override with `CHROME_DEBUGGER_PORT`); if unreachable, Chrome is
  killed. Blocked tabs are redirected to `/tmp/freedxm-block.html`, written
  from the embedded `chrome/1.html`.
- `process/` — reads process names from `/proc/[pid]/stat`; kills with
  `killall -TERM`, then `-9`, deduplicating concurrent kills of the same name.
- `notifier/` — runs user-configured shell commands (`%title`, `%message`
  placeholders) and always logs the notification.
- `config/` — reads TOML from `~/.config/freedxm.toml` (override with
  `CONFIG_FILEPATH`). The config is re-read on demand, not cached, so edits
  take effect without restarting the server.
- `util/` — small generic helpers (set/slice, TOML/JSON, `SecondsToHHMMSS`,
  `IdempotentRunner` which skips `Run` if already running via `TryLock`).

Session changes flow through channels: the gRPC server signals
`sessionsUpdateCh` → `listenSessionsUpdated` recomputes matchers → notifies
observers (the Chrome monitor) with non-blocking sends. The design tolerates
brief inconsistency that the next poll cycle resolves; see the comment at the
top of `commands/serve.go`.

## Conventions

- Lines ≤ 80 characters (enforced by revive, except tests). Wrap long
  function signatures one parameter per line, as existing code does.
- revive has nearly all rules enabled: avoid magic numbers (name them as
  constants; `0, 1, 2, 100` are allowed), check returned errors, keep
  cognitive complexity low.
- Exported identifiers do not need doc comments; comments are sparse and
  explain *why*.
- Use the standard `log` package for logging; return errors rather than
  exiting outside `main`.
- Tests use the standard `testing` package with table-driven cases, placed
  next to the code (`*_test.go`).
- Commit messages: short, lowercase, imperative, optionally prefixed with an
  area (e.g. `ls command: ...`, `readme: ...`).
