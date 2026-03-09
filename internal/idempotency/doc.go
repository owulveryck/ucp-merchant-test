// Package idempotency implements idempotency key handling for the UCP Shopping
// Service REST transport.
//
// The Universal Commerce Protocol requires businesses to support idempotency
// for state-changing operations (checkout creation, updates, completion). When
// a platform retries a request with the same Idempotency-Key header, the
// business must return the same response without re-executing the operation.
// This prevents duplicate orders and ensures safe retries in the face of
// network failures.
//
// The Store tracks previously seen idempotency keys along with their request
// payload hash, HTTP status code, and response body. On a subsequent request
// with the same key:
//
//   - If the payload hash matches, the cached response is returned directly.
//   - If the payload hash differs, the business rejects the request since
//     reusing an idempotency key with different content indicates a client error.
//
// The HashPayload function computes a SHA-256 digest of the request body for
// comparison purposes. The Reset method clears all stored entries, used by
// the conformance test suite's simulation reset endpoint.
package idempotency
