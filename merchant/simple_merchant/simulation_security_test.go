package main

import (
	"testing"
)

func TestSimulationEndpointMissingHeader(t *testing.T) {
	ts := newTestServer(t)
	orderID := ts.createCompletedOrder()

	headers := ts.getHeaders("")
	resp, _ := ts.doRequest("POST", "/testing/simulate-shipping/"+orderID, nil, headers)
	if resp.StatusCode != 403 {
		t.Fatalf("Expected 403, got %d", resp.StatusCode)
	}
}

func TestSimulationEndpointIncorrectSecret(t *testing.T) {
	ts := newTestServer(t)
	orderID := ts.createCompletedOrder()

	headers := ts.getHeaders("")
	headers["Simulation-Secret"] = "for-sure-incorrect-secret"
	resp, _ := ts.doRequest("POST", "/testing/simulate-shipping/"+orderID, nil, headers)
	if resp.StatusCode != 403 {
		t.Fatalf("Expected 403, got %d", resp.StatusCode)
	}
}

func TestSimulationEndpointCorrectSecret(t *testing.T) {
	ts := newTestServer(t)
	orderID := ts.createCompletedOrder()

	headers := ts.getHeaders("")
	headers["Simulation-Secret"] = simulationSecret
	resp, _ := ts.doRequest("POST", "/testing/simulate-shipping/"+orderID, nil, headers)
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
}
