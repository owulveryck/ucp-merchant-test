package rest

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
)

func TestGetOrder_Success(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.GetOrderFunc = func(id, ownerID string) (*model.Order, error) {
		if id != "ord_1" {
			t.Errorf("expected id ord_1, got %s", id)
		}
		return &model.Order{ID: "ord_1", CheckoutID: "co_1"}, nil
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/orders/ord_1", nil)
	injectAuth(t, req, authSrv, "user1", "US")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var order model.Order
	json.NewDecoder(resp.Body).Decode(&order)
	if order.ID != "ord_1" {
		t.Errorf("expected order id ord_1, got %s", order.ID)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	ts, mock, authSrv := setupTestServer(t)

	mock.GetOrderFunc = func(id, ownerID string) (*model.Order, error) {
		return nil, merchant.ErrNotFound
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/orders/nonexistent", nil)
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

func TestUpdateOrder_Success(t *testing.T) {
	ts, mock, _ := setupTestServer(t)

	mock.UpdateOrderFunc = func(id string, req model.OrderUpdateRequest) (*model.Order, error) {
		if id != "ord_1" {
			t.Errorf("expected id ord_1, got %s", id)
		}
		return &model.Order{ID: "ord_1"}, nil
	}

	body := `{"fulfillment":{"events":[]}}`
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/orders/ord_1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUpdateOrder_InvalidJSON(t *testing.T) {
	ts, _, _ := setupTestServer(t)

	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/orders/ord_1", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}
