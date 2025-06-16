package browser

import (
	"os"
	"testing"
)

func TestLoadStorage(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "storage")
	if err != nil {
		t.Fatal(err)
	}
	tmp.WriteString("hello")
	tmp.Close()
	data, err := LoadStorage(tmp.Name())
	if err != nil || string(data) != "hello" {
		t.Fatalf("unexpected result: %v %s", err, data)
	}
}
