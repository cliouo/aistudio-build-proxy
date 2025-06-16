package browser

import (
	"os"
	"testing"
)

func TestLoadStorage(t *testing.T) {
	path := t.TempDir() + "/state.json"
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	p, err := LoadStorage(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != path {
		t.Fatalf("expected %s, got %s", path, p)
	}
}
