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
- **Throughput (local, Grade A)** — **2026-09-30** re-run: **10** clients × **1,000** cmds (**30%** writes) → **78,086 ops/sec**, P95 **0.21 ms**. Artifact: [`docs/evidence/bench-2026-09-30.txt`](docs/evidence/bench-2026-09-30.txt). (Prior example **74,498** / **0.20 ms** remains in [docs/benchmark.md](docs/benchmark.md) as Grade C history.)
- **Correctness** — **11** Go tests (`parser` + `store`); CI runs `go test ./…` ([`docs/evidence/gotest-2026-09-30.txt`](docs/evidence/gotest-2026-09-30.txt)).

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
| **78,086 ops/sec**, P95 **0.21 ms** (2026-09-30) | [`docs/evidence/bench-2026-09-30.txt`](docs/evidence/bench-2026-09-30.txt) — Grade **A** |
| **74,498 ops/sec**, P95 **0.20 ms** (prior) | `docs/benchmark.md` — Grade **C** history |
| **11** tests green | [`docs/evidence/gotest-2026-09-30.txt`](docs/evidence/gotest-2026-09-30.txt) |
| AOF replay | `aof.go` + server startup path |

Do not pitch multi-node Redis cluster parity — this is a single-process learning/systems build with a measured local bench.

---

## For interview depth

| Doc | Use |
|-----|-----|
| [docs/METRICS.md](docs/METRICS.md) | Claim ↔ evidence grades |
| [docs/DECISIONS.md](docs/DECISIONS.md) | Why AOF / RWMutex / bench honesty |
| [docs/benchmark.md](docs/benchmark.md) | Bench method |
