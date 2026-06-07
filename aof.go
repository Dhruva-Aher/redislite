package main

import (
	"io"
	"os"
	"sync"
)

// AOF (Append-Only File) handles persistence
// Instead of complex snapshots, we just append every write command to a file.
type AOF struct {
	file *os.File
	mu   sync.Mutex
}

func NewAOF(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &AOF{file: f}, nil
}

func (a *AOF) Write(val Value) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, err := a.file.Write(val.Marshal())
	return err
}

func (a *AOF) Read(callback func(value Value)) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Start reading from the beginning
	a.file.Seek(0, 0)
	
	parser := NewParser(a.file)

	for {
		val, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		callback(val)
	}

	// Move back to the end so future writes append correctly
	a.file.Seek(0, 2)
	return nil
}

func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Close()
}
