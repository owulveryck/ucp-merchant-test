package main

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// IdempotencyEntry stores the result of a previously processed request.
type IdempotencyEntry struct {
	PayloadHash  string
	StatusCode   int
	ResponseBody []byte
}

var (
	idempotencyStore = map[string]*IdempotencyEntry{}
	idempotencyMu    sync.Mutex
)

func hashPayload(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// checkIdempotency checks if a request with this key was already processed.
// Returns (entry, exists).
func checkIdempotency(key string) (*IdempotencyEntry, bool) {
	idempotencyMu.Lock()
	defer idempotencyMu.Unlock()
	entry, ok := idempotencyStore[key]
	return entry, ok
}

// storeIdempotency records a processed request.
func storeIdempotency(key, payloadHash string, statusCode int, body []byte) {
	idempotencyMu.Lock()
	defer idempotencyMu.Unlock()
	idempotencyStore[key] = &IdempotencyEntry{
		PayloadHash:  payloadHash,
		StatusCode:   statusCode,
		ResponseBody: body,
	}
}
