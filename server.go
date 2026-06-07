package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

// Server holds our server state
type Server struct {
	port  string
	store *Store
}

func NewServer(port string) *Server {
	return &Server{
		port:  port,
		store: NewStore(),
	}
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

		switch command {
		case "PING":
			conn.Write(Value{Type: "string", Str: "PONG"}.Marshal())
		case "SET":
			if len(val.Array) < 3 {
				conn.Write(Value{Type: "error", Str: "ERR wrong number of arguments for 'set' command"}.Marshal())
				continue
			}
			s.store.Set(val.Array[1].Str, val.Array[2].Str)
			conn.Write(Value{Type: "string", Str: "OK"}.Marshal())
		case "GET":
			if len(val.Array) != 2 {
				conn.Write(Value{Type: "error", Str: "ERR wrong number of arguments for 'get' command"}.Marshal())
				continue
			}
			res, ok := s.store.Get(val.Array[1].Str)
			if !ok {
				conn.Write(Value{Type: "bulk", IsNull: true}.Marshal())
			} else {
				conn.Write(Value{Type: "bulk", Str: res}.Marshal())
			}
		case "DEL":
			if len(val.Array) != 2 {
				conn.Write(Value{Type: "error", Str: "ERR wrong number of arguments for 'del' command"}.Marshal())
				continue
			}
			count := s.store.Del(val.Array[1].Str)
			conn.Write(Value{Type: "integer", Num: count}.Marshal())
		case "EXPIRE":
			if len(val.Array) != 3 {
				conn.Write(Value{Type: "error", Str: "ERR wrong number of arguments for 'expire' command"}.Marshal())
				continue
			}
			ttl, err := strconv.Atoi(val.Array[2].Str)
			if err != nil {
				conn.Write(Value{Type: "error", Str: "ERR value is not an integer or out of range"}.Marshal())
				continue
			}
			ok := s.store.Expire(val.Array[1].Str, ttl)
			if ok {
				conn.Write(Value{Type: "integer", Num: 1}.Marshal())
			} else {
				conn.Write(Value{Type: "integer", Num: 0}.Marshal())
			}
		case "TTL":
			if len(val.Array) != 2 {
				conn.Write(Value{Type: "error", Str: "ERR wrong number of arguments for 'ttl' command"}.Marshal())
				continue
			}
			res := s.store.TTL(val.Array[1].Str)
			conn.Write(Value{Type: "integer", Num: res}.Marshal())
		case "HSET":
			if len(val.Array) < 4 {
				conn.Write(Value{Type: "error", Str: "ERR wrong number of arguments for 'hset' command"}.Marshal())
				continue
			}
			s.store.HSet(val.Array[1].Str, val.Array[2].Str, val.Array[3].Str)
			conn.Write(Value{Type: "integer", Num: 1}.Marshal())
		case "HGET":
			if len(val.Array) != 3 {
				conn.Write(Value{Type: "error", Str: "ERR wrong number of arguments for 'hget' command"}.Marshal())
				continue
			}
			res, ok := s.store.HGet(val.Array[1].Str, val.Array[2].Str)
			if !ok {
				conn.Write(Value{Type: "bulk", IsNull: true}.Marshal())
			} else {
				conn.Write(Value{Type: "bulk", Str: res}.Marshal())
			}
		case "LPUSH":
			if len(val.Array) < 3 {
				conn.Write(Value{Type: "error", Str: "ERR wrong number of arguments for 'lpush' command"}.Marshal())
				continue
			}
			var args []string
			for i := 2; i < len(val.Array); i++ {
				args = append(args, val.Array[i].Str)
			}
			count := s.store.LPush(val.Array[1].Str, args...)
			conn.Write(Value{Type: "integer", Num: count}.Marshal())
		case "LRANGE":
			if len(val.Array) != 4 {
				conn.Write(Value{Type: "error", Str: "ERR wrong number of arguments for 'lrange' command"}.Marshal())
				continue
			}
			start, err1 := strconv.Atoi(val.Array[2].Str)
			stop, err2 := strconv.Atoi(val.Array[3].Str)
			if err1 != nil || err2 != nil {
				conn.Write(Value{Type: "error", Str: "ERR value is not an integer or out of range"}.Marshal())
				continue
			}
			res := s.store.LRange(val.Array[1].Str, start, stop)
			arr := Value{Type: "array", Array: make([]Value, 0)}
			for _, item := range res {
				arr.Array = append(arr.Array, Value{Type: "bulk", Str: item})
			}
			conn.Write(arr.Marshal())
		default:
			conn.Write(Value{Type: "error", Str: "ERR unknown command '" + command + "'"}.Marshal())
		}
	}
}
