package teal

import (
	"os"
	"path"
	"testing"
)

func TestParser(t *testing.T) {
	fs, err := os.ReadDir("examples")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range fs {
		if path.Ext(f.Name()) != ".teal" {
			continue
		}

		bs, err := os.ReadFile(path.Join("examples", f.Name()))
		if err != nil {
			t.Fatal(err)
		}

		_, err = Parse(string(bs))
		if err != nil {
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
