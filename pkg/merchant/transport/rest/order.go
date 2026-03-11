package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/owulveryck/ucp-merchant-test/pkg/model"
)

// handleOrders is the main router for /orders/.
func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getOrder(w, r)
	case http.MethodPut:
		s.updateOrder(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	id := extractPathParam(r.URL.Path, "/orders/")
	ownerID := s.auth.ExtractUserFromToken(r)

	order, err := s.merchant.GetOrder(id, ownerID)
	if mapError(w, err) {
		return
	}

	writeJSONResponse(w, http.StatusOK, order)
}

func (s *Server) updateOrder(w http.ResponseWriter, r *http.Request) {
	id := extractPathParam(r.URL.Path, "/orders/")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	var req model.OrderUpdateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Invalid JSON")
		return
	}

	// Validate adjustment statuses
	validStatuses := map[string]bool{"pending": true, "approved": true, "rejected": true, "completed": true}
	for _, adj := range req.Adjustments {
		if adj.Status != "" && !validStatuses[adj.Status] {
			writeError(w, http.StatusUnprocessableEntity, "Invalid adjustment status: "+adj.Status)
			return
		}
	}

	order, err := s.merchant.UpdateOrder(id, req)
	if mapError(w, err) {
		return
	}

	writeJSONResponse(w, http.StatusOK, order)
}
