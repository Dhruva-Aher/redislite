# RedisLite

I built this to understand how Redis actually works under the hood. It's a Redis-compatible key-value store written in Go from scratch. It speaks a subset of the actual RESP protocol (the wire protocol Redis uses), which means you can connect to it using the standard `redis-cli`!

This is a personal project to learn about TCP servers, concurrent programming with goroutines, and writing parsers. 

## What works

- TCP server listens on port 6380
- Parses basic RESP protocol using a custom parser (no libs!)
- Responds to `PING` with `+PONG`
- Goroutines handle each connection concurrently.
- Core in-memory data store using `sync.RWMutex`
- Supported commands: `SET`, `GET`, `DEL`, `EXPIRE`, `TTL`, `HSET`, `HGET`, `LPUSH`, `LRANGE`, `INFO`
- Background goroutine that evicts expired keys every 100ms (basically what Redis does under the hood)
- Append-Only File (AOF) persistence: every write command is logged to `redislite.aof` and replayed on startup

## Architecture Flow

```text
redis-cli
    │
    ▼
TCP Server
    │
RESP Parser
    │
Command Handler
    │
Store (RWMutex)
    │
AOF Log
```

## How to run

1. Make sure you have Go installed.
2. Run `go build` in this directory.
3. Start the server with `./redislite`.
4. Open another terminal and connect using the standard Redis CLI on port 6380:
   `redis-cli -p 6380`
5. Try out some commands!

## Benchmarks

I wrote a standalone benchmarking tool (`benchmark.go`) to test throughput and latency using pure stdlib TCP connections. It spawns concurrent goroutines that send a configurable mix of SET and GET commands.

**How to run it:**
```bash
go run benchmark.go -clients 10 -commands 1000 -mix 30
```

**Results on my machine:**
```text
Results:
  clients:     10
  commands:    10000
  duration:    0.13s
  throughput:  74,498 ops/sec
  p95 latency: 0.20ms
```

To scale this further, the single `sync.RWMutex` over the entire data store would need to be sharded (array of maps) to reduce lock contention under heavy concurrent writes.

## How it works

- **RESP (REdis Serialization Protocol):** Redis clients and servers talk to each other using a surprisingly simple text-based protocol. Commands are sent as arrays of "bulk strings" (strings with their length prefixed). The server parses these arrays byte-by-byte using a custom parser and sends back standard RESP responses. RESP is weirdly simple once you get it.
- **Goroutines:** When a client connects, the server spins up a new goroutine to handle their session. This means RedisLite can handle multiple clients typing commands simultaneously without blocking each other. Go makes concurrency super straightforward.
- **AOF (Append-Only File):** Instead of saving complex snapshots of memory, RedisLite just writes every single modification command (like `SET` or `DEL`) to a file called `redislite.aof`. When the server restarts, it literally just replays those same commands from the file to reconstruct its previous state.

## What I learned

Building this helped me realize that a lot of database magic is actually just basic data structures and good locking strategies. Writing a protocol parser from scratch sounded terrifying but it's really just reading bytes from a buffer. I also learned that `sync.RWMutex` is great when you have a lot of `GET` commands, though honestly, having one giant lock for the whole store isn't going to scale to millions of requests (TODO: optimize locking later). 

One thing that's still missing: error handling for edge cases where clients send malformed or partial frames, but it works fine for standard `redis-cli` usage!
