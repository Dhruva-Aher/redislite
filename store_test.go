package main

import (
	"testing"
	"time"
)

func TestStore_SetAndGet(t *testing.T) {
	s := NewStore()
	
	s.Set("key1", "val1")
	val, ok := s.Get("key1")
	
	if !ok {
		t.Errorf("Expected key1 to exist")
	}
	if val != "val1" {
		t.Errorf("Expected val1, got %s", val)
	}
}

func TestStore_GetNonExistent(t *testing.T) {
	s := NewStore()
	
	_, ok := s.Get("missing")
	if ok {
		t.Errorf("Expected missing key to return false")
	}
}

func TestStore_Delete(t *testing.T) {
	s := NewStore()
	
	s.Set("key1", "val1")
	count := s.Del("key1")
	
	if count != 1 {
		t.Errorf("Expected delete count to be 1, got %d", count)
	}
	
	_, ok := s.Get("key1")
	if ok {
		t.Errorf("Expected key1 to be deleted")
	}
}

func TestStore_ExpireAndTTL(t *testing.T) {
	s := NewStore()
	
	s.Set("key1", "val1")
	s.Expire("key1", 1) // 1 second expiry
	
	ttl := s.TTL("key1")
	if ttl < 0 || ttl > 1 {
		t.Errorf("Expected TTL to be 0 or 1, got %d", ttl)
	}
	
	// Wait for expiry
	time.Sleep(1100 * time.Millisecond)
	
	_, ok := s.Get("key1")
	if ok {
		t.Errorf("Expected key1 to be expired")
	}
	
	ttlAfter := s.TTL("key1")
	if ttlAfter != -2 {
		t.Errorf("Expected TTL of expired key to be -2, got %d", ttlAfter)
	}
}

func TestStore_Hashes(t *testing.T) {
	s := NewStore()
	
	s.HSet("user:1", "name", "alice")
	s.HSet("user:1", "age", "30")
	
	val, ok := s.HGet("user:1", "name")
	if !ok || val != "alice" {
		t.Errorf("Expected name to be alice, got %s", val)
	}
	
	val, ok = s.HGet("user:1", "missing_field")
	if ok {
		t.Errorf("Expected missing_field to return false")
	}
}

func TestStore_Lists(t *testing.T) {
	s := NewStore()
	
	// LPUSH acts a bit differently in our simplified clone (bulk append)
	// but we should still test the LRange bounds logic
	s.LPush("list1", "a", "b", "c")
	
	res := s.LRange("list1", 0, -1)
	if len(res) != 3 {
		t.Errorf("Expected list length 3, got %d", len(res))
	}
	
	if res[0] != "a" || res[2] != "c" {
		t.Errorf("Expected [a b c], got %v", res)
	}
	
	// Test partial range
	res = s.LRange("list1", 0, 1)
	if len(res) != 2 {
		t.Errorf("Expected list length 2, got %d", len(res))
	}
}
