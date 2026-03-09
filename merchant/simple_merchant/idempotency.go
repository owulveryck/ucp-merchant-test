package main

import (
	"github.com/owulveryck/ucp-merchant-test/internal/idempotency"
)

// IdempotencyEntry is an alias for backward compatibility.
type IdempotencyEntry = idempotency.Entry

// Global idempotency store instance.
var idempotencyStoreInstance = idempotency.NewStore()

func hashPayload(body []byte) string {
	return idempotency.HashPayload(body)
}

func checkIdempotency(key string) (*IdempotencyEntry, bool) {
	return idempotencyStoreInstance.Check(key)
}

func storeIdempotency(key, payloadHash string, statusCode int, body []byte) {
	idempotencyStoreInstance.Store(key, payloadHash, statusCode, body)
}
