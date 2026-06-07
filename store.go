package main

import (
	"sync"
	"time"
)

// Item holds our value and an optional expiry time
type Item struct {
	Value      string
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
	s.data[key] = Item{Value: value}
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
	return item.Value, true
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
