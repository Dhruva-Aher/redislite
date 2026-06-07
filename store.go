package main

import (
	"sync"
	"time"
)

// Item holds our value and an optional expiry time
type Item struct {
	Type       string // "string", "hash", "list"
	StrVal     string
	HashVal    map[string]string
	ListVal    []string
	Expiration int64 // unix nano, 0 if no expiry
}

// Store is the main in-memory data store.
// We use a sync.RWMutex because GETs usually outnumber SETs.
type Store struct {
	mu   sync.RWMutex
	data map[string]Item
}

func NewStore() *Store {
	s := &Store{
		data: make(map[string]Item),
	}
	// background eviction goroutine
	// this is basically what Redis does under the hood (well, kinda)
	go s.evictExpired()
	return s
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = Item{Type: "string", StrVal: value}
}

func (s *Store) Expire(key string, ttlSeconds int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[key]
	if !ok {
		return false
	}
	// update expiration
	item.Expiration = time.Now().UnixNano() + (int64(ttlSeconds) * int64(time.Second))
	s.data[key] = item
	return true
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.data[key]
	if !ok {
		return "", false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return "", false // expired but not yet evicted
	}
	if item.Type != "string" {
		return "", false // wrong type
	}
	return item.StrVal, true
}

func (s *Store) HSet(key, field, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[key]
	if !ok || item.Type != "hash" {
		item = Item{Type: "hash", HashVal: make(map[string]string)}
	}
	item.HashVal[field] = value
	s.data[key] = item
}

func (s *Store) HGet(key, field string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.data[key]
	if !ok || item.Type != "hash" {
		return "", false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return "", false
	}
	val, ok := item.HashVal[field]
	return val, ok
}

func (s *Store) LPush(key string, values ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[key]
	if !ok || item.Type != "list" {
		item = Item{Type: "list", ListVal: make([]string, 0)}
	}
	// LPUSH adds to the head (beginning)
	item.ListVal = append(values, item.ListVal...)
	s.data[key] = item
	return len(item.ListVal)
}

func (s *Store) LRange(key string, start, stop int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.data[key]
	if !ok || item.Type != "list" {
		return nil
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil
	}

	listLen := len(item.ListVal)
	if listLen == 0 {
		return nil
	}

	// Handle negative indices
	if start < 0 {
		start = listLen + start
	}
	if stop < 0 {
		stop = listLen + stop
	}
	if start < 0 {
		start = 0
	}
	if stop < 0 {
		stop = 0
	}
	if start >= listLen {
		return nil
	}
	if stop >= listLen {
		stop = listLen - 1
	}
	if start > stop {
		return nil
	}

	return item.ListVal[start : stop+1]
}

func (s *Store) Del(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		delete(s.data, key)
		return 1
	}
	return 0
}

func (s *Store) TTL(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.data[key]
	if !ok {
		return -2 // key does not exist
	}
	if item.Expiration == 0 {
		return -1 // exists but no ttl
	}
	now := time.Now().UnixNano()
	if now > item.Expiration {
		return -2 // expired
	}
	// return TTL in seconds
	return int((item.Expiration - now) / int64(time.Second))
}

// evictExpired runs every 100ms and cleans up expired keys
func (s *Store) evictExpired() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for range ticker.C {
		now := time.Now().UnixNano()
		
		// TODO: add LRU eviction later or optimize this because 
		// locking the whole map every 100ms for large datasets is bad
		s.mu.Lock()
		for k, v := range s.data {
			if v.Expiration > 0 && now > v.Expiration {
				delete(s.data, k)
			}
		}
		s.mu.Unlock()
	}
}
