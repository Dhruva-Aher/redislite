package main

import (
	"bytes"
	"testing"
)

func TestParser_ParseArray(t *testing.T) {
	input := "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n"
	parser := NewParser(bytes.NewBufferString(input))
	
	val, err := parser.Parse()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if val.Type != "array" {
		t.Fatalf("Expected array type, got %s", val.Type)
	}
	
	if len(val.Array) != 3 {
		t.Fatalf("Expected array of length 3, got %d", len(val.Array))
	}
	
	if val.Array[0].Str != "SET" {
		t.Errorf("Expected SET, got %s", val.Array[0].Str)
	}
	
	if val.Array[2].Str != "myvalue" {
		t.Errorf("Expected myvalue, got %s", val.Array[2].Str)
	}
}

func TestParser_NullBulkString(t *testing.T) {
	input := "$-1\r\n"
	parser := NewParser(bytes.NewBufferString(input))
	
	val, err := parser.Parse()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if val.Type != "bulk" || !val.IsNull {
		t.Errorf("Expected null bulk string")
	}
}

func TestMarshal_String(t *testing.T) {
	val := Value{Type: "string", Str: "OK"}
	res := string(val.Marshal())
	
	expected := "+OK\r\n"
	if res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}

func TestMarshal_Integer(t *testing.T) {
	val := Value{Type: "integer", Num: 42}
	res := string(val.Marshal())
	
	expected := ":42\r\n"
	if res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}

func TestMarshal_BulkString(t *testing.T) {
	val := Value{Type: "bulk", Str: "hello"}
	res := string(val.Marshal())
	
	expected := "$5\r\nhello\r\n"
	if res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}
