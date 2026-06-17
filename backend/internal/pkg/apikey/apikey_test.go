package apikey

import "testing"

func TestGenerateHashAndMask(t *testing.T) {
	key, err := Generate("lsk_live_")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(key) <= len("lsk_live_") {
		t.Fatalf("key too short: %q", key)
	}
	if Hash(key) == Hash(key+"x") {
		t.Fatal("hash collision for different keys")
	}
	if Mask(key) == key {
		t.Fatal("mask should not expose full key")
	}
}
