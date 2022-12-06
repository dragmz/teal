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

		errs := Lint(string(bs))
		for _, err := range errs {
			t.Errorf("failed to parse - file: %s, error: %s", f.Name(), err)
		}
	}
}

func TestReadHexInt(t *testing.T) {
	tests := []struct {
		In  string
		Out uint64
	}{
		{In: "0x01", Out: 1},
		{In: "0x0001", Out: 256},
		{In: "0xff", Out: 255},
		{In: "0xFF", Out: 255},
	}

	for _, test := range tests {
		v, err := readHexInt(test.In)
		if err != nil {
			t.Error(err)
		}

		if v != test.Out {
			t.Errorf("unexpected value: %d != 1", v)
		}
	}
}
