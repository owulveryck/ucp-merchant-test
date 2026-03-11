package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/webhook"
)

// handleCheckoutSessions is the main router for /shopping-api/checkout-sessions.
func (s *Server) handleCheckoutSessions(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/shopping-api/checkout-sessions")
	path = strings.TrimSuffix(path, "/")

	switch {
	case path == "" || path == "/":
		if r.Method == http.MethodPost {
			s.createCheckout(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	case strings.HasSuffix(path, "/complete"):
		if r.Method == http.MethodPost {
			s.completeCheckout(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	case strings.HasSuffix(path, "/cancel"):
		if r.Method == http.MethodPost {
			s.cancelCheckout(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	default:
		switch r.Method {
		case http.MethodGet:
			s.getCheckout(w, r)
		case http.MethodPut:
			s.updateCheckout(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// mapError converts a merchant error to an HTTP status code and writes the error response.
// Returns true if an error was written.
func mapError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, merchant.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, merchant.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, merchant.ErrBadRequest):
		status = http.StatusBadRequest
	case errors.Is(err, merchant.ErrPaymentFailed):
		status = http.StatusPaymentRequired
	case errors.Is(err, merchant.ErrForbidden):
		status = http.StatusForbidden
	}
	writeError(w, status, err.Error())
	return true
}

func (s *Server) createCheckout(w http.ResponseWriter, r *http.Request) {
	if checkVersionNegotiation(w, r) {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	if s.handleIdempotency(w, r, body) {
		return
	}

	var req model.CheckoutRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	ownerID := s.auth.ExtractUserFromToken(r)
	country := s.auth.ExtractUserCountry(r)

	co, _, err := s.merchant.CreateCheckout(ownerID, country, &req)
	if mapError(w, err) {
		return
	}

	// Store webhook URL from UCP-Agent header
	webhookURL := webhook.ResolveWebhookURL(r.Header.Get("UCP-Agent"))
	if webhookURL != "" {
		s.SetWebhookURL(co.ID, webhookURL)
	}

	s.processAndRespond(w, r, body, http.StatusCreated, co)
}

func (s *Server) getCheckout(w http.ResponseWriter, r *http.Request) {
	id := extractPathParam(r.URL.Path, "/shopping-api/checkout-sessions/")
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}

	ownerID := s.auth.ExtractUserFromToken(r)

	co, _, err := s.merchant.GetCheckout(id, ownerID)
	if mapError(w, err) {
		return
	}

	writeJSONResponse(w, http.StatusOK, co)
}

func (s *Server) updateCheckout(w http.ResponseWriter, r *http.Request) {
	if checkVersionNegotiation(w, r) {
		return
	}

	id := extractPathParam(r.URL.Path, "/shopping-api/checkout-sessions/")
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	if s.handleIdempotency(w, r, body) {
		return
	}

	var req model.CheckoutRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	ownerID := s.auth.ExtractUserFromToken(r)

	co, _, err := s.merchant.UpdateCheckout(id, ownerID, &req)
	if mapError(w, err) {
		return
	}

	s.processAndRespond(w, r, body, http.StatusOK, co)
}

func (s *Server) completeCheckout(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	prefix := "/shopping-api/checkout-sessions/"
	suffix := "/complete"
	id := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	if s.handleIdempotency(w, r, body) {
		return
	}

	var req model.CheckoutRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	ownerID := s.auth.ExtractUserFromToken(r)
	country := s.auth.ExtractUserCountry(r)

	// REST flow: empty approvalHash skips hash validation
	co, order, _, err := s.merchant.CompleteCheckout(id, ownerID, country, "", &req)
	if mapError(w, err) {
		return
	}

	// Send webhook
	if webhookURL := s.GetWebhookURL(id); webhookURL != "" {
		webhook.SendWebhookEvent(webhookURL, model.WebhookEvent{
			EventType:  "order_placed",
			CheckoutID: id,
			Order:      order,
		})
	}

	s.processAndRespond(w, r, body, http.StatusOK, co)
}

func (s *Server) cancelCheckout(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	prefix := "/shopping-api/checkout-sessions/"
	suffix := "/cancel"
	id := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)

	body, _ := io.ReadAll(r.Body)

	if s.handleIdempotency(w, r, body) {
		return
	}

	ownerID := s.auth.ExtractUserFromToken(r)

	co, _, err := s.merchant.CancelCheckout(id, ownerID)
	if mapError(w, err) {
		return
	}

	s.processAndRespond(w, r, body, http.StatusOK, co)
}
