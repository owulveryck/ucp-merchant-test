package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// Entry stores the result of a previously processed request.
type Entry struct {
	PayloadHash  string
	StatusCode   int
	ResponseBody []byte
}

// Store provides idempotent request tracking.
type Store struct {
	mu      sync.Mutex
	entries map[string]*Entry
}

// NewStore creates a new idempotency store.
func NewStore() *Store {
	return &Store{
		entries: map[string]*Entry{},
	}
}

// HashPayload computes a SHA-256 hash of the request body.
func HashPayload(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// Check checks if a request with this key was already processed.
func (s *Store) Check(key string) (*Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[key]
	return entry, ok
}

// Store records a processed request.
func (s *Store) Store(key, payloadHash string, statusCode int, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = &Entry{
		PayloadHash:  payloadHash,
		StatusCode:   statusCode,
		ResponseBody: body,
	}
}

// Reset clears all stored entries.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = map[string]*Entry{}
}
