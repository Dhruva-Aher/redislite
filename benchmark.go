//go:build ignore
// +build ignore

// This script measures the throughput and p95 latency of RedisLite.
// It connects over real TCP and runs concurrent clients sending a mix
// of SET and GET commands to see how the server holds up under load.
//
// How to run:
//   go run benchmark.go -clients 10 -commands 1000 -mix 30

package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"
)

func main() {
	clients := flag.Int("clients", 10, "number of concurrent clients")
	commands := flag.Int("commands", 1000, "number of commands per client")
	host := flag.String("host", "localhost:6380", "server address")
	mix := flag.Int("mix", 30, "read/write ratio as write percentage 0-100")

	flag.Parse()

	fmt.Printf("Starting benchmark: %d clients, %d commands each, %d%% writes\n", *clients, *commands, *mix)

	var wg sync.WaitGroup
	// Channel to collect all latencies
	totalCommands := (*clients) * (*commands)
	latencies := make([]time.Duration, totalCommands)

	start := time.Now()

	for i := 0; i < *clients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", *host)
			if err != nil {
				fmt.Println("Error connecting to server:", err)
				return
			}
			defer conn.Close()

			rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(clientID)))

			for j := 0; j < *commands; j++ {
				key := fmt.Sprintf("benchkey:%d:%d", clientID, rng.Intn(100))
				
				var cmd string
				if rng.Intn(100) < *mix {
					// SET command
					val := "benchval"
					cmd = fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(val), val)
				} else {
					// GET command
					cmd = fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
				}

				t0 := time.Now()
				rw.WriteString(cmd)
				rw.Flush()

				// Read response
				line, _ := rw.ReadString('\n')
				if len(line) > 0 && line[0] == '$' {
					if line != "$-1\r\n" {
						rw.ReadString('\n') // read the bulk string value
					}
				}

				lat := time.Since(t0)
				latencies[(clientID * *commands) + j] = lat
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Calculate metrics
	var total time.Duration
	for _, l := range latencies {
		total += l
	}

	throughput := float64(totalCommands) / duration.Seconds()

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p95Idx := int(float64(totalCommands) * 0.95)
	if p95Idx >= totalCommands {
		p95Idx = totalCommands - 1
	}
	p95 := latencies[p95Idx]

	fmt.Println("\nResults:")
	fmt.Printf("  clients:     %d\n", *clients)
	fmt.Printf("  commands:    %d\n", totalCommands)
	fmt.Printf("  duration:    %.2fs\n", duration.Seconds())
	// Use fmt.Printf for thousands separator (trick: there is no simple one in stdlib fmt, so we just print it as %.0f or format it manually, but the user wanted "8,064 ops/sec". Let's do a simple manual format if needed, or just %.0f is fine without commas to stick to stdlib without extra logic, but wait, I can write a quick thousands formatter).
	fmt.Printf("  throughput:  %s ops/sec\n", formatThousands(int(throughput)))
	fmt.Printf("  p95 latency: %.2fms\n", float64(p95)/float64(time.Millisecond))
}

// formatThousands adds commas to an integer
func formatThousands(n int) string {
	in := fmt.Sprintf("%d", n)
	out := make([]byte, len(in)+(len(in)-1)/3)
	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			break
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
	return string(out)
}
