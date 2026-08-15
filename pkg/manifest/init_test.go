package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteStarter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, DefaultPath)
	if err := WriteStarter(path, false); err != nil {
		t.Fatal(err)
	}
	if err := WriteStarter(path, false); err == nil {
		t.Fatal("expected refuse overwrite")
	}
	if err := WriteStarter(path, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Services["example"]; !ok {
		t.Fatalf("services=%v", m.Services)
	}
}
