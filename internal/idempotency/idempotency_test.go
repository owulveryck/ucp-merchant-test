package idempotency

import "testing"

func TestStoreAndCheck(t *testing.T) {
	s := NewStore()

	_, exists := s.Check("key1")
	if exists {
		t.Error("expected no entry for key1")
	}

	s.Store("key1", "hash1", 200, []byte(`{"ok":true}`))

	entry, exists := s.Check("key1")
	if !exists {
		t.Fatal("expected entry for key1")
	}
	if entry.PayloadHash != "hash1" {
		t.Errorf("expected hash1, got %s", entry.PayloadHash)
	}
	if entry.StatusCode != 200 {
		t.Errorf("expected 200, got %d", entry.StatusCode)
	}
}

func TestReset(t *testing.T) {
	s := NewStore()
	s.Store("key1", "hash1", 200, []byte(`{}`))
	s.Reset()

	_, exists := s.Check("key1")
	if exists {
		t.Error("expected no entry after reset")
	}
}

func TestHashPayload(t *testing.T) {
	h := HashPayload([]byte("test"))
	if h == "" {
		t.Error("expected non-empty hash")
	}
	h2 := HashPayload([]byte("test"))
	if h != h2 {
		t.Error("expected same hash for same input")
	}
	h3 := HashPayload([]byte("other"))
	if h == h3 {
		t.Error("expected different hash for different input")
	}
}
