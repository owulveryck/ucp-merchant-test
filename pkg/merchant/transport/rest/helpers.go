package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/owulveryck/ucp-merchant-test/pkg/idempotency"
)

func writeJSONResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, detail string) {
	writeJSONResponse(w, status, map[string]string{"detail": detail})
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")
}

// checkVersionNegotiation checks the UCP-Agent header for version compatibility.
// Returns true if the request should be rejected.
func checkVersionNegotiation(w http.ResponseWriter, r *http.Request) bool {
	ucpAgent := r.Header.Get("UCP-Agent")
	if ucpAgent == "" {
		return false
	}
	for _, part := range strings.Split(ucpAgent, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "version=") {
			version := strings.Trim(strings.TrimPrefix(part, "version="), "\"")
			if version != "" && version != "2026-01-11" {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("Incompatible UCP version: %s. Expected 2026-01-11", version))
				return true
			}
		}
	}
	return false
}

func (s *Server) handleIdempotency(w http.ResponseWriter, r *http.Request, body []byte) bool {
	if s.idempotency == nil {
		return false
	}
	key := r.Header.Get("idempotency-key")
	if key == "" {
		return false
	}
	payloadHash := idempotency.HashPayload(body)
	entry, exists := s.idempotency.Check(key)
	if !exists {
		return false
	}
	if entry.PayloadHash == payloadHash {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(entry.StatusCode)
		w.Write(entry.ResponseBody)
		return true
	}
	writeError(w, http.StatusConflict, "Idempotency key conflict: payload differs from original request")
	return true
}

func (s *Server) storeIdempotentResponse(r *http.Request, body []byte, statusCode int, responseBody []byte) {
	if s.idempotency == nil {
		return
	}
	key := r.Header.Get("idempotency-key")
	if key == "" {
		return
	}
	s.idempotency.Store(key, idempotency.HashPayload(body), statusCode, responseBody)
}

func (s *Server) processAndRespond(w http.ResponseWriter, r *http.Request, reqBody []byte, status int, result interface{}) {
	respBody, _ := json.Marshal(result)
	s.storeIdempotentResponse(r, reqBody, status, respBody)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(respBody)
}

func extractPathParam(path, prefix string) string {
	s := strings.TrimPrefix(path, prefix)
	s = strings.TrimSuffix(s, "/")
	return s
}
