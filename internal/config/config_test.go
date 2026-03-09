package config

import "testing"

func TestScheme(t *testing.T) {
	c := &Config{TLSEnabled: false}
	if c.Scheme() != "http" {
		t.Errorf("expected http, got %s", c.Scheme())
	}
	c.TLSEnabled = true
	if c.Scheme() != "https" {
		t.Errorf("expected https, got %s", c.Scheme())
	}
}
