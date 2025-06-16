package browser

import (
	"os"
	"testing"
)

func TestResolveEndpoint(t *testing.T) {
	os.WriteFile("/tmp/endpoint.txt", []byte("ws://bar"), 0o644)
	defer os.Remove("/tmp/endpoint.txt")

	if ep := ResolveEndpoint("ws://explicit"); ep != "ws://explicit" {
		t.Fatalf("unexpected %s", ep)
	}
	if ep := ResolveEndpoint(""); ep != "ws://bar" {
		t.Fatalf("unexpected %s", ep)
	}
}
