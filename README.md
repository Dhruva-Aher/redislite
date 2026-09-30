# RedisLite

**Redis-compatible KV store** · Systems · Go · TCP · RESP

From-scratch in-memory store that speaks a Redis RESP subset over raw TCP — concurrent connections, TTL eviction, and AOF persistence — so `redis-cli` works without a Redis dependency.

[![CI](https://github.com/Dhruva-Aher/redislite/actions/workflows/ci.yml/badge.svg)](https://github.com/Dhruva-Aher/redislite/actions/workflows/ci.yml)

| | |
|--|--|
| **Focus** | Wire protocol · concurrency · persistence |
| **Stack** | Go · TCP · RESP · `sync.RWMutex` · AOF |
| **Proof** | `go test ./…` · in-repo TCP benchmark |

---

## Highlights

- **Protocol** — Custom RESP parser (no Redis client libs); `PING` / `SET` / `GET` / `DEL` / `EXPIRE` / `TTL` / hashes / lists / `INFO`.
- **Concurrency** — Per-connection goroutines; shared store guarded by `sync.RWMutex`; background TTL sweep every **100 ms**.
- **Durability** — Append-only file (`redislite.aof`) logs writes and replays on startup.
- **Throughput (local)** — Documented bench: **10** clients × **1,000** cmds (**30%** writes) → **74,498 ops/sec**, P95 **0.20 ms** (M-series Mac; see [docs/benchmark.md](docs/benchmark.md)).
- **Correctness** — **11** Go tests (`parser` + `store`); CI runs `go test ./…`.

---

## Architecture

| Component | Responsibility |
|-----------|----------------|
| **TCP server** | Accept on `:6380`, one goroutine per client |
| **RESP parser** | Decode requests / encode replies |
| **Command handler** | Dispatch Redis-subset commands |
| **Store** | In-memory maps + RWMutex + TTL |
| **AOF** | Persist writes; replay on boot |

```text
redis-cli -p 6380
    → TCP server → RESP parser → command handler
                              → Store (RWMutex) → AOF log
```

Depth: [docs/architecture.md](docs/architecture.md) · [docs/protocol.md](docs/protocol.md) · [docs/benchmark.md](docs/benchmark.md)

---

## Quick start

```bash
go build -o redislite .
./redislite
# other terminal
redis-cli -p 6380
PING
SET foo bar
GET foo
```

Tests: `go test ./…`  
Bench (server running): `go run . -bench` or see [docs/benchmark.md](docs/benchmark.md)

---

## Evidence notes

| Claim | Evidence |
|-------|----------|
| **74,498 ops/sec**, P95 **0.20 ms** | Example run in `docs/benchmark.md` (local; not a multi-machine capacity claim) |
| **11** tests | `parser_test.go` + `store_test.go` |
| AOF replay | `aof.go` + server startup path |

Do not pitch multi-node Redis cluster parity — this is a single-process learning/systems build with a measured local bench.
