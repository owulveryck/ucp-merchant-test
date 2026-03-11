package rest

import (
	"net/http"
	"testing"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func TestSimulateShipping_Success(t *testing.T) {
	ts, mock, _ := setupTestServer(t)

	mock.UpdateOrderFunc = func(id string, req model.OrderUpdateRequest) (*model.Order, error) {
		return &model.Order{
			ID: id,
			LineItems: []model.OrderLineItem{
				{ID: "li_1", Quantity: model.OrderQuantity{Total: 2}},
			},
		}, nil
	}

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/testing/simulate-shipping/ord_1", nil)
	req.Header.Set("Simulation-Secret", "test-secret")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestSimulateShipping_WrongSecret(t *testing.T) {
	ts, _, _ := setupTestServer(t)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/testing/simulate-shipping/ord_1", nil)
	req.Header.Set("Simulation-Secret", "wrong-secret")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestSimulateShipping_MissingSecret(t *testing.T) {
	ts, _, _ := setupTestServer(t)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/testing/simulate-shipping/ord_1", nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}
