package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

// Server holds our server state
type Server struct {
	port string
}

func NewServer(port string) *Server {
	return &Server{port: port}
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
		default:
			conn.Write(Value{Type: "error", Str: "ERR unknown command '" + command + "'"}.Marshal())
		}
	}
}
