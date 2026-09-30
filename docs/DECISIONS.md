# Decisions — RedisLite

Status: **PROPOSED** ≠ **DECIDED** ≠ **IMPLEMENTED** ≠ **VERIFIED**  
Related: [METRICS.md](./METRICS.md) · [benchmark.md](./benchmark.md) · [architecture.md](./architecture.md) · [protocol.md](./protocol.md)

**Cross-verify (2026-09-30):** Test count **11** green. Throughput upgraded to Grade **A** — **78,086 ops/sec** / P95 **0.21 ms** in `docs/evidence/bench-2026-09-30.txt`.

---

## D1 — Build a Redis subset from scratch (not a client wrapper)

| | |
|--|--|
| **Context** | Need a systems portfolio piece: TCP, concurrency, wire protocol. |
| **Decision** | Implement RESP over raw TCP + in-memory store in Go; speak enough Redis for `redis-cli`. |
| **Why** | Interview depth on parser + goroutines + locking, not “used go-redis”. |
| **Alternatives** | Fork miniredis; wrap real Redis. |
| **Tradeoffs** | Incomplete Redis surface; single-node only. |
| **Evidence** | `parser.go`, `server.go`, `store.go` |
| **Status** | DECIDED · IMPLEMENTED |

---

## D2 — Global `sync.RWMutex` over the keyspace

| | |
|--|--|
| **Context** | Concurrent clients share one map. |
| **Decision** | One RWMutex guarding the store (reads concurrent, writes exclusive). |
| **Why** | Correct and simple for teaching/portfolio scale. |
| **Alternatives** | Shard by key hash (256 mutexes); channel-serialized single writer. |
| **Tradeoffs** | Write throughput ceiling under contention — noted in benchmark.md. |
| **Evidence** | `store.go`; scaling notes in `docs/benchmark.md` |
| **Status** | DECIDED · IMPLEMENTED |

---

## D3 — AOF for durability (not RDB snapshots)

| | |
|--|--|
| **Context** | Need persistence story without Redis RDB complexity. |
| **Decision** | Append every write as RESP to `redislite.aof`; replay on boot. |
| **Why** | Matches Redis mental model; easy to explain crash recovery. |
| **Alternatives** | Periodic gob snapshot; no persistence. |
| **Tradeoffs** | Unbounded AOF growth; no rewrite compaction yet. |
| **Evidence** | `aof.go` |
| **Status** | DECIDED · IMPLEMENTED |

---

## D4 — Public throughput from archived local bench only

| | |
|--|--|
| **Context** | Temptation to invent “100k+ QPS” marketing. |
| **Decision** | Prefer Grade A file under `docs/evidence/`; keep older `docs/benchmark.md` example as Grade C history. |
| **Why** | Portfolio honesty; FAANG interviewers ask for method. |
| **Evidence** | `docs/evidence/bench-2026-09-30.txt` (**78,086** ops/sec) |
| **Status** | DECIDED · VERIFIED |

---

## D5 — CI = `go test` only

| | |
|--|--|
| **Context** | Bench needs a running server; flaky in CI without orchestration. |
| **Decision** | GitHub Actions runs unit tests; bench stays manual/docs. |
| **Evidence** | `.github/workflows/ci.yml` |
| **Status** | DECIDED · IMPLEMENTED |
