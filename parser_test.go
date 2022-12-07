package teal

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestParser(t *testing.T) {
	p := filepath.Join("examples", "ok")
	fs, err := os.ReadDir(p)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range fs {
		if path.Ext(f.Name()) != ".teal" {
			continue
		}

		bs, err := os.ReadFile(path.Join(p, f.Name()))
		if err != nil {
			t.Fatal(err)
		}

		res := Process(string(bs))
		for _, d := range res.Diagnostics {
			t.Errorf("failed to parse - file: %s, error: %s", f.Name(), d)
		}
	}
}
