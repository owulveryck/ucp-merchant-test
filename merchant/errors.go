// Sentinel errors returned by Merchant interface methods.
//
// Transport packages use errors.Is() to map these to protocol-specific
// error codes:
//   - REST: ErrNotFound->404, ErrConflict->409, ErrBadRequest->400,
//     ErrPaymentFailed->402, ErrForbidden->403
//   - MCP: all errors -> MCPToolResult with IsError=true
package merchant

import "errors"

// ErrNotFound indicates the requested resource does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict indicates a state conflict (e.g., completing an already
// completed checkout session).
var ErrConflict = errors.New("conflict")

// ErrBadRequest indicates invalid input from the caller.
var ErrBadRequest = errors.New("bad request")

// ErrPaymentFailed indicates the payment credential was rejected
// during checkout completion.
var ErrPaymentFailed = errors.New("payment failed")

// ErrForbidden indicates the caller lacks permission to access
// the requested resource (ownerID mismatch).
var ErrForbidden = errors.New("forbidden")
