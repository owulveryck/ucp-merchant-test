package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/owulveryck/ucp-merchant-test/pkg/auth"
	"github.com/owulveryck/ucp-merchant-test/pkg/idempotency"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/merchanttest"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/ucp"
)

func setupTestServer(t *testing.T) (*httptest.Server, *merchanttest.Mock, *auth.OAuthServer) {
	t.Helper()
	mock := merchanttest.NewMock()
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	srv := New(mock, authSrv,
		WithSimulationSecret("test-secret"),
		WithIdempotency(idempotency.NewStore()),
	)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, mock, authSrv
}

func injectAuth(t *testing.T, req *http.Request, authSrv *auth.OAuthServer, userID string, country ucp.Country) {
	t.Helper()
	token := authSrv.InjectToken(userID, country, time.Now().Add(time.Hour))
	req.Header.Set("Authorization", "Bearer "+token)
}

func TestRouting_CheckoutPOST(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	called := false
	mock.CreateCheckoutFunc = func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		called = true
		return &model.Checkout{ID: "co_1", Status: "incomplete"}, "", nil
	}

	body := `{"line_items":[{"product_id":"SKU-001","quantity":1}]}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	injectAuth(t, req, authSrv, "user1", "US")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
	if !called {
		t.Error("CreateCheckout was not called")
	}
}

func TestRouting_CheckoutGET(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.GetCheckoutFunc = func(id, ownerID string) (*model.Checkout, string, error) {
		if id != "co_1" {
			t.Errorf("expected id co_1, got %s", id)
		}
		return &model.Checkout{ID: "co_1", Status: "incomplete"}, "", nil
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/shopping-api/checkout-sessions/co_1", nil)
	injectAuth(t, req, authSrv, "user1", "US")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRouting_MethodNotAllowed(t *testing.T) {
	ts, _, _ := setupTestServer(t)

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/shopping-api/checkout-sessions", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestReset_ClearsWebhookURLs(t *testing.T) {
	mock := merchanttest.NewMock()
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	srv := New(mock, authSrv, WithIdempotency(idempotency.NewStore()))

	srv.SetWebhookURL("co_1", "http://example.com/webhook")
	if got := srv.GetWebhookURL("co_1"); got != "http://example.com/webhook" {
		t.Fatalf("expected webhook URL before reset, got %q", got)
	}

	srv.Reset()

	if got := srv.GetWebhookURL("co_1"); got != "" {
		t.Errorf("expected empty after reset, got %q", got)
	}
}

func TestIdempotency_CachedResponse(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.CreateCheckoutFunc = func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		return &model.Checkout{ID: "co_1", Status: "incomplete"}, "", nil
	}

	body := `{"line_items":[{"product_id":"SKU-001","quantity":1}]}`

	// First request
	req1, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("idempotency-key", "key-1")
	injectAuth(t, req1, authSrv, "user1", "US")

	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()

	// Second request with same key+payload
	mock.CreateCheckoutFunc = func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		return nil, "", merchant.ErrBadRequest // Should not be called
	}

	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("idempotency-key", "key-1")
	injectAuth(t, req2, authSrv, "user1", "US")

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusCreated {
		t.Errorf("expected cached 201, got %d", resp2.StatusCode)
	}
}

func TestIdempotency_ConflictOnDifferentPayload(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.CreateCheckoutFunc = func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		return &model.Checkout{ID: "co_1", Status: "incomplete"}, "", nil
	}

	body1 := `{"line_items":[{"product_id":"SKU-001","quantity":1}]}`
	req1, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions", strings.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("idempotency-key", "key-2")
	injectAuth(t, req1, authSrv, "user1", "US")

	resp1, _ := http.DefaultClient.Do(req1)
	resp1.Body.Close()

	body2 := `{"line_items":[{"product_id":"SKU-002","quantity":2}]}`
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("idempotency-key", "key-2")
	injectAuth(t, req2, authSrv, "user1", "US")

	resp2, _ := http.DefaultClient.Do(req2)
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("expected 409, got %d", resp2.StatusCode)
	}

	var errBody map[string]string
	json.Unmarshal(readBody(t, resp2), &errBody)
	if !strings.Contains(errBody["detail"], "Idempotency") {
		t.Error("expected idempotency conflict message")
	}
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
