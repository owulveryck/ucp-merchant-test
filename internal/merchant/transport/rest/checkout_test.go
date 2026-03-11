package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/owulveryck/ucp-merchant-test/internal/merchant"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

func TestCreateCheckout_Success(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	var gotOwnerID string
	var gotCountry ucp.Country
	mock.CreateCheckoutFunc = func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		gotOwnerID = ownerID
		gotCountry = country
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
	if gotOwnerID != "user1" {
		t.Errorf("expected ownerID user1, got %s", gotOwnerID)
	}
	if gotCountry != "US" {
		t.Errorf("expected country US, got %s", gotCountry)
	}
}

func TestCreateCheckout_InvalidJSON(t *testing.T) {
	ts, _, authSrv := setupTestServer(t)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	injectAuth(t, req, authSrv, "user1", "US")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateCheckout_MerchantBadRequest(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.CreateCheckoutFunc = func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		return nil, "", fmt.Errorf("invalid item: %w", merchant.ErrBadRequest)
	}

	body := `{"line_items":[{"product_id":"INVALID","quantity":1}]}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	injectAuth(t, req, authSrv, "user1", "US")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateCheckout_VersionNegotiation(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	called := false
	mock.CreateCheckoutFunc = func(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		called = true
		return &model.Checkout{ID: "co_1"}, "", nil
	}

	body := `{"line_items":[{"product_id":"SKU-001","quantity":1}]}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("UCP-Agent", "test/1.0; version=9999-01-01")
	injectAuth(t, req, authSrv, "user1", "US")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if called {
		t.Error("merchant should not have been called with wrong version")
	}
}

func TestGetCheckout_Success(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.GetCheckoutFunc = func(id, ownerID string) (*model.Checkout, string, error) {
		if id != "co_42" {
			t.Errorf("expected id co_42, got %s", id)
		}
		return &model.Checkout{ID: "co_42", Status: "incomplete"}, "", nil
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/shopping-api/checkout-sessions/co_42", nil)
	injectAuth(t, req, authSrv, "user1", "US")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var co model.Checkout
	json.NewDecoder(resp.Body).Decode(&co)
	if co.ID != "co_42" {
		t.Errorf("expected id co_42, got %s", co.ID)
	}
}

func TestGetCheckout_NotFound(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.GetCheckoutFunc = func(id, ownerID string) (*model.Checkout, string, error) {
		return nil, "", merchant.ErrNotFound
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/shopping-api/checkout-sessions/nonexistent", nil)
	injectAuth(t, req, authSrv, "user1", "US")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdateCheckout_Success(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.UpdateCheckoutFunc = func(id, ownerID string, req *model.CheckoutRequest) (*model.Checkout, string, error) {
		return &model.Checkout{ID: id, Status: "incomplete"}, "", nil
	}

	body := `{"line_items":[{"product_id":"SKU-002","quantity":2}]}`
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/shopping-api/checkout-sessions/co_1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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

func TestCompleteCheckout_Success(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	var gotApprovalHash string
	mock.CompleteCheckoutFunc = func(id, ownerID string, country ucp.Country, approvalHash string, req *model.CheckoutRequest) (*model.Checkout, *model.Order, string, error) {
		gotApprovalHash = approvalHash
		return &model.Checkout{ID: id, Status: "completed"}, &model.Order{ID: "ord_1"}, "", nil
	}

	body := `{"payment":{"token":"success_token"}}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions/co_1/complete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	injectAuth(t, req, authSrv, "user1", "US")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if gotApprovalHash != "" {
		t.Errorf("REST should pass empty approvalHash, got %q", gotApprovalHash)
	}
}

func TestCompleteCheckout_PaymentFailed(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.CompleteCheckoutFunc = func(id, ownerID string, country ucp.Country, approvalHash string, req *model.CheckoutRequest) (*model.Checkout, *model.Order, string, error) {
		return nil, nil, "", merchant.ErrPaymentFailed
	}

	body := `{"payment":{"token":"fail_token"}}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions/co_1/complete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	injectAuth(t, req, authSrv, "user1", "US")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPaymentRequired {
		t.Errorf("expected 402, got %d", resp.StatusCode)
	}
}

func TestCancelCheckout_Success(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.CancelCheckoutFunc = func(id, ownerID string) (*model.Checkout, string, error) {
		return &model.Checkout{ID: id, Status: "canceled"}, "", nil
	}

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/shopping-api/checkout-sessions/co_1/cancel", nil)
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
