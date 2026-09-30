# Metrics & Claims — RedisLite

**Cross-verified:** 2026-09-30 (bench re-run archived)

| ID | Claim (exact) | Grade | Evidence |
|----|---------------|-------|----------|
| C1 | TCP `:6380` RESP subset; `redis-cli` compatible | A | `server.go`, `parser.go` |
| C2 | Commands: PING, SET, GET, DEL, EXPIRE, TTL, hashes, lists, INFO | A | handlers |
| C3 | TTL eviction every **100 ms** | A | background sweep |
| C4 | AOF persistence + replay | A | `aof.go` |
| C5 | **11** Go tests green | A | [`docs/evidence/gotest-2026-09-30.txt`](./evidence/gotest-2026-09-30.txt) |
| C6 | Bench **78,086 ops/sec**, P95 **0.21 ms** (10×1000, 30% writes) | A | [`docs/evidence/bench-2026-09-30.txt`](./evidence/bench-2026-09-30.txt) |
| C7 | Prior example **74,498 ops/sec**, P95 **0.20 ms** | C | `docs/benchmark.md` |
| C8 | CI `go test ./…` | A | `.github/workflows/ci.yml` |

## Re-verify

```bash
go test ./... -count=1 | tee docs/evidence/gotest-$(date -u +%Y-%m-%d).txt
go build -o redislite . && ./redislite &
go run benchmark.go -clients 10 -commands 1000 -mix 30 | tee docs/evidence/bench-$(date -u +%Y-%m-%d).txt
```
