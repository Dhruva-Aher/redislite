# RedisLite

I built this to understand how Redis actually works under the hood. It's a Redis-compatible key-value store written in Go from scratch. It speaks a subset of the actual RESP protocol (the wire protocol Redis uses), which means you can connect to it using the standard `redis-cli`!

This is a personal project to learn about TCP servers, concurrent programming with goroutines, and writing parsers. 

## What works so far

Phase 3 is complete! 
- TCP server listens on port 6380
- Parses basic RESP protocol using a custom parser (no libs!)
- Responds to `PING` with `+PONG`
- Handles concurrent connections using goroutines (which really makes this almost too easy).
- Core in-memory data store using `sync.RWMutex`
- Supported commands: `SET`, `GET`, `DEL`, `EXPIRE`, `TTL`
- Background goroutine that evicts expired keys every 100ms (basically what Redis does under the hood)
- Added support for Hashes (`HSET`, `HGET`) and Lists (`LPUSH`, `LRANGE`)

## Supported Commands

- `PING`
- `SET`
- `GET`
- `DEL`
- `EXPIRE`
- `TTL`
- `HSET`
- `HGET`
- `LPUSH`
- `LRANGE`
(To be implemented)
- `PING`
- `INFO`

## How to run

(Instructions coming soon!)

## How it works

(Coming soon!)

## What I learned

(I'll write this after I finish!)
