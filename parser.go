package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// Value represents a RESP value
type Value struct {
	Type   string
	Str    string
	Num    int
	Array  []Value
	IsNull bool
}

// RESP is weirdly simple once you get it.
// We just need a bufio.Reader to read byte by byte or line by line.

type Parser struct {
	reader *bufio.Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(r)}
}

// Parse reads the next RESP value from the reader
func (p *Parser) Parse() (Value, error) {
	b, err := p.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch b {
	case '*':
		return p.parseArray()
	case '$':
		return p.parseBulkString()
	// TODO: handle simple strings (+), integers (:), errors (-) later if needed
	// For now, clients like redis-cli send arrays of bulk strings for commands
	default:
		// Not strictly RESP, or we're out of sync
		return Value{}, fmt.Errorf("unknown RESP type: %q", b)
	}
}

// readLine reads until \r\n and returns the line without the CRLF
func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) >= 2 && line[len(line)-2] == '\r' {
		return line[:len(line)-2], nil
	}
	return line[:len(line)-1], nil
}

func (p *Parser) parseArray() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}

	length, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, err
	}

	var arr []Value
	for i := 0; i < length; i++ {
		val, err := p.Parse()
		if err != nil {
			return Value{}, err
		}
		arr = append(arr, val)
	}

	return Value{Type: "array", Array: arr}, nil
}

func (p *Parser) parseBulkString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}

	length, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, err
	}

	if length == -1 {
		return Value{Type: "bulk", IsNull: true}, nil
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(p.reader, buf)
	if err != nil {
		return Value{}, err
	}

	// Read the trailing \r\n
	_, err = p.readLine()
	if err != nil {
		return Value{}, err
	}

	return Value{Type: "bulk", Str: string(buf)}, nil
}

// Marshal converts a Value back to a RESP string
func (v Value) Marshal() []byte {
	switch v.Type {
	case "string":
		return []byte(fmt.Sprintf("+%s\r\n", v.Str))
	case "error":
		return []byte(fmt.Sprintf("-%s\r\n", v.Str))
	case "integer":
		return []byte(fmt.Sprintf(":%d\r\n", v.Num))
	case "bulk":
		if v.IsNull {
			return []byte("$-1\r\n")
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v.Str), v.Str))
	case "array":
		res := []byte(fmt.Sprintf("*%d\r\n", len(v.Array)))
		for _, item := range v.Array {
			res = append(res, item.Marshal()...)
		}
		return res
	default:
		return []byte("")
	}
}
