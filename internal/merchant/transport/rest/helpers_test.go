package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/owulveryck/ucp-merchant-test/internal/merchant"
)

func TestMapError_Nil(t *testing.T) {
	w := httptest.NewRecorder()
	if mapError(w, nil) {
		t.Error("mapError(nil) should return false")
	}
	if w.Code != http.StatusOK {
		t.Errorf("no status should have been written, got %d", w.Code)
	}
}

func TestMapError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	if !mapError(w, fmt.Errorf("item: %w", merchant.ErrNotFound)) {
		t.Error("mapError should return true for ErrNotFound")
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["detail"] == "" {
		t.Error("expected detail in error body")
	}
}

func TestMapError_Conflict(t *testing.T) {
	w := httptest.NewRecorder()
	mapError(w, merchant.ErrConflict)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestMapError_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	mapError(w, merchant.ErrBadRequest)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestMapError_PaymentFailed(t *testing.T) {
	w := httptest.NewRecorder()
	mapError(w, merchant.ErrPaymentFailed)
	if w.Code != http.StatusPaymentRequired {
		t.Errorf("expected 402, got %d", w.Code)
	}
}

func TestMapError_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	mapError(w, merchant.ErrForbidden)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestMapError_Unknown(t *testing.T) {
	w := httptest.NewRecorder()
	mapError(w, errors.New("something"))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestCheckVersionNegotiation_NoHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	if checkVersionNegotiation(w, r) {
		t.Error("should return false when no UCP-Agent header")
	}
}

func TestCheckVersionNegotiation_CorrectVersion(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("UCP-Agent", "platform/1.0; version=2026-01-11")
	if checkVersionNegotiation(w, r) {
		t.Error("should return false for correct version")
	}
}

func TestCheckVersionNegotiation_WrongVersion(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("UCP-Agent", "platform/1.0; version=9999-01-01")
	if !checkVersionNegotiation(w, r) {
		t.Error("should return true for wrong version")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestExtractPathParam(t *testing.T) {
	tests := []struct {
		path, prefix, want string
	}{
		{"/orders/ord_123", "/orders/", "ord_123"},
		{"/orders/ord_123/", "/orders/", "ord_123"},
		{"/shopping-api/checkout-sessions/co_1", "/shopping-api/checkout-sessions/", "co_1"},
	}
	for _, tt := range tests {
		got := extractPathParam(tt.path, tt.prefix)
		if got != tt.want {
			t.Errorf("extractPathParam(%q, %q) = %q, want %q", tt.path, tt.prefix, got, tt.want)
		}
	}
}

func TestWriteJSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSONResponse(w, http.StatusCreated, map[string]string{"id": "123"})

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["id"] != "123" {
		t.Errorf("expected id=123, got %s", body["id"])
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusBadRequest, "bad input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["detail"] != "bad input" {
		t.Errorf("expected detail='bad input', got %q", body["detail"])
	}
}
