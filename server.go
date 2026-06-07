package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Server holds our server state
type Server struct {
	port      string
	store     *Store
	aof       *AOF
	clients   int
	commands  int
	startTime time.Time
	statsMu   sync.Mutex
}

func NewServer(port string) *Server {
	aof, err := NewAOF("redislite.aof")
	if err != nil {
		fmt.Println("Warning: could not initialize AOF:", err)
	}

	server := &Server{
		port:      port,
		store:     NewStore(),
		aof:       aof,
		startTime: time.Now(),
	}

	if aof != nil {
		server.loadAOF()
	}

	return server
}

func (s *Server) loadAOF() {
	fmt.Println("Replaying AOF file...")
	s.aof.Read(func(val Value) {
		if val.Type != "array" || len(val.Array) == 0 {
			return
		}
		command := strings.ToUpper(val.Array[0].Str)
		s.handleCommand(command, val, nil)
	})
	fmt.Println("AOF replay complete")
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("RedisLite server listening on %s\n", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		// goroutines make this almost too easy
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	s.statsMu.Lock()
	s.clients++
	s.statsMu.Unlock()

	defer func() {
		s.statsMu.Lock()
		s.clients--
		s.statsMu.Unlock()
	}()

	fmt.Println("New client connected:", conn.RemoteAddr())

	parser := NewParser(conn)

	for {
		val, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected:", conn.RemoteAddr())
			} else {
				fmt.Println("Error parsing RESP:", err)
			}
			return
		}

		// We expect commands as an array of bulk strings
		if val.Type != "array" || len(val.Array) == 0 {
			continue
		}

		// Extract the command name (the first string in the array)
		cmdVal := val.Array[0]
		if cmdVal.Type != "bulk" {
			continue
		}
		command := strings.ToUpper(cmdVal.Str)

		s.statsMu.Lock()
		s.commands++
		s.statsMu.Unlock()

		s.handleCommand(command, val, conn)
	}
}

func (s *Server) handleCommand(command string, val Value, conn net.Conn) {
	// Helper to conditionally write to conn
	writeConn := func(v Value) {
		if conn != nil {
			conn.Write(v.Marshal())
		}
	}

	// Helper to conditionally append to AOF
	writeAOF := func() {
		if s.aof != nil && conn != nil {
			s.aof.Write(val)
		}
	}

	switch command {
	case "PING":
		writeConn(Value{Type: "string", Str: "PONG"})
	case "INFO":
		s.statsMu.Lock()
		clients := s.clients
		commands := s.commands
		s.statsMu.Unlock()
		uptime := time.Since(s.startTime).Seconds()

		infoStr := fmt.Sprintf("# Server\r\nredislite_version:1.0.0\r\nuptime_in_seconds:%d\r\n\r\n# Clients\r\nconnected_clients:%d\r\n\r\n# Stats\r\ntotal_commands_processed:%d\r\n", int(uptime), clients, commands)
		writeConn(Value{Type: "bulk", Str: infoStr})
	case "SET":
		if len(val.Array) < 3 {
			writeConn(Value{Type: "error", Str: "ERR wrong number of arguments for 'set' command"})
			return
		}
		s.store.Set(val.Array[1].Str, val.Array[2].Str)
		writeAOF()
		writeConn(Value{Type: "string", Str: "OK"})
	case "GET":
		if len(val.Array) != 2 {
			writeConn(Value{Type: "error", Str: "ERR wrong number of arguments for 'get' command"})
			return
		}
		res, ok := s.store.Get(val.Array[1].Str)
		if !ok {
			writeConn(Value{Type: "bulk", IsNull: true})
		} else {
			writeConn(Value{Type: "bulk", Str: res})
		}
	case "DEL":
		if len(val.Array) != 2 {
			writeConn(Value{Type: "error", Str: "ERR wrong number of arguments for 'del' command"})
			return
		}
		count := s.store.Del(val.Array[1].Str)
		writeAOF()
		writeConn(Value{Type: "integer", Num: count})
	case "EXPIRE":
		if len(val.Array) != 3 {
			writeConn(Value{Type: "error", Str: "ERR wrong number of arguments for 'expire' command"})
			return
		}
		ttl, err := strconv.Atoi(val.Array[2].Str)
		if err != nil {
			writeConn(Value{Type: "error", Str: "ERR value is not an integer or out of range"})
			return
		}
		ok := s.store.Expire(val.Array[1].Str, ttl)
		if ok {
			writeAOF()
			writeConn(Value{Type: "integer", Num: 1})
		} else {
			writeConn(Value{Type: "integer", Num: 0})
		}
	case "TTL":
		if len(val.Array) != 2 {
			writeConn(Value{Type: "error", Str: "ERR wrong number of arguments for 'ttl' command"})
			return
		}
		res := s.store.TTL(val.Array[1].Str)
		writeConn(Value{Type: "integer", Num: res})
	case "HSET":
		if len(val.Array) < 4 {
			writeConn(Value{Type: "error", Str: "ERR wrong number of arguments for 'hset' command"})
			return
		}
		s.store.HSet(val.Array[1].Str, val.Array[2].Str, val.Array[3].Str)
		writeAOF()
		writeConn(Value{Type: "integer", Num: 1})
	case "HGET":
		if len(val.Array) != 3 {
			writeConn(Value{Type: "error", Str: "ERR wrong number of arguments for 'hget' command"})
			return
		}
		res, ok := s.store.HGet(val.Array[1].Str, val.Array[2].Str)
		if !ok {
			writeConn(Value{Type: "bulk", IsNull: true})
		} else {
			writeConn(Value{Type: "bulk", Str: res})
		}
	case "LPUSH":
		if len(val.Array) < 3 {
			writeConn(Value{Type: "error", Str: "ERR wrong number of arguments for 'lpush' command"})
			return
		}
		var args []string
		for i := 2; i < len(val.Array); i++ {
			args = append(args, val.Array[i].Str)
		}
		count := s.store.LPush(val.Array[1].Str, args...)
		writeAOF()
		writeConn(Value{Type: "integer", Num: count})
	case "LRANGE":
		if len(val.Array) != 4 {
			writeConn(Value{Type: "error", Str: "ERR wrong number of arguments for 'lrange' command"})
			return
		}
		start, err1 := strconv.Atoi(val.Array[2].Str)
		stop, err2 := strconv.Atoi(val.Array[3].Str)
		if err1 != nil || err2 != nil {
			writeConn(Value{Type: "error", Str: "ERR value is not an integer or out of range"})
			return
		}
		res := s.store.LRange(val.Array[1].Str, start, stop)
		arr := Value{Type: "array", Array: make([]Value, 0)}
		for _, item := range res {
			arr.Array = append(arr.Array, Value{Type: "bulk", Str: item})
		}
		writeConn(arr)
	default:
		writeConn(Value{Type: "error", Str: "ERR unknown command '" + command + "'"})
	}
}
