package model

import "testing"

func TestUCPEnvelopeDefaults(t *testing.T) {
	env := UCPEnvelope{Version: "2026-01-11", Capabilities: []Capability{}}
	if env.Version != "2026-01-11" {
		t.Errorf("expected version 2026-01-11, got %s", env.Version)
	}
	if len(env.Capabilities) != 0 {
		t.Errorf("expected empty capabilities, got %d", len(env.Capabilities))
	}
}

func TestCheckoutStatus(t *testing.T) {
	co := Checkout{ID: "co_001", Status: "incomplete"}
	if co.Status != "incomplete" {
		t.Errorf("expected incomplete, got %s", co.Status)
	}
}
