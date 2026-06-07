# The Benchmark

The `benchmark.go` script is a custom-built, standalone testing tool designed to measure the throughput (ops/sec) and latency of RedisLite.

## How it works

The benchmark connects directly to RedisLite over real TCP sockets, exactly like a real application would. 

1. **Concurrency**: It spins up N concurrent goroutines (simulating N independent clients).
2. **Workload**: Each goroutine sends M commands in a tight loop. It uses a realistic read/write ratio (default is 70% `GET`, 30% `SET`).
3. **Accuracy**: 
   - Commands are fully serialized into raw RESP strings.
   - The script waits for, reads, and validates the full server response before marking a command as complete. `GET`s do not fail silently.
   - We measure latency natively using `time.Since()` for every individual command.

## Methodology

Because we aren't using a third-party testing framework, all statistics are calculated manually:
- **Throughput**: Total number of commands executed divided by total wall-clock duration.
- **P95 Latency**: Every single command's latency is recorded into a massive array. When the benchmark finishes, we sort the array and pluck the value exactly at the 95th percentile index.

## Example Results

On a standard M-series Mac with the default configuration (`10 clients, 1000 commands each, 30% writes`), the system comfortably achieves:

```text
Results:
  clients:     10
  commands:    10000
  duration:    0.13s
  throughput:  74,498 ops/sec
  p95 latency: 0.20ms
```

## Scaling Considerations

Currently, 75k ops/sec is excellent for a single local instance. However, if we needed to scale beyond this, the primary bottleneck is the global `sync.RWMutex` locking the single overarching data `map`.

To scale writes further, we would implement **lock sharding**. Instead of one map, we would create an array of 256 maps, each with its own mutex. We would hash the incoming key to determine which shard it belongs to, reducing lock contention significantly.
