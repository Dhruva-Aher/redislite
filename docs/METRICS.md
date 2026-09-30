# Metrics & Claims — RedisLite

**Cross-verified:** 2026-09-30  
Evidence grades: **A** = artifact/command this tree · **B** = reproducible harness · **C** = historical documented run · **D** = design target only

## Claim table

| ID | Claim (exact) | Grade | Evidence |
|----|---------------|-------|----------|
| C1 | TCP server on **:6380**; RESP subset; `redis-cli` compatible | A | `server.go`, `parser.go`, README quick start |
| C2 | Commands: PING, SET, GET, DEL, EXPIRE, TTL, HSET, HGET, LPUSH, LRANGE, INFO | A | command handler in `server.go` / store |
| C3 | TTL eviction every **100 ms** | A | background goroutine in server/store |
| C4 | AOF persistence + replay on startup | A | `aof.go` |
| C5 | **11** Go tests (`parser` **5** + `store` **6**) | A | `go test ./…` → ok (2026-09-30) |
| C6 | Bench: **10** clients × **1000** cmds, **30%** writes → **74,498 ops/sec**, P95 **0.20 ms** | C | Example in `docs/benchmark.md` (M-series Mac). Re-run locally with `go run benchmark.go` against a live server for Grade A refresh. |
| C7 | CI runs `go test ./…` | A | `.github/workflows/ci.yml` |

## Explicit non-claims

| Phrase | Why |
|--------|-----|
| Redis Cluster / multi-node | Single-process map + RWMutex |
| Sustained production QPS | One local example run only |
| “3.5M ops/sec” or any failed-connect bench | Invalid — clients must complete RESP round-trips |

## How to re-verify

```bash
go test ./... -count=1
go build -o redislite . && ./redislite &
go run benchmark.go -clients 10 -commands 1000 -mix 30
```
