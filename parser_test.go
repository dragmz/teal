package teal

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestParser(t *testing.T) {
	type test struct {
		Path  string
		Clean bool
	}

	tests := []test{
		{
			Path:  filepath.Join("examples", "ok"),
			Clean: true,
		},
		{
			Path:  filepath.Join("examples", "err"),
			Clean: false,
		},
	}

	for _, test := range tests {
		fs, err := os.ReadDir(test.Path)
		if err != nil {
			t.Fatal(err)
		}

		for _, f := range fs {
			if path.Ext(f.Name()) != ".teal" {
				continue
			}

			bs, err := os.ReadFile(path.Join(test.Path, f.Name()))
			if err != nil {
				t.Fatal(err)
			}

			res := Process(string(bs))
			if test.Clean {
				for _, d := range res.Diagnostics {
					t.Errorf("failed to parse - file: %s, error: %s", f.Name(), d)
				}
			} else {
				if len(res.Diagnostics) == 0 {
					t.Errorf("expected errors but got none: %s", test.Path)
				}
			}
		}
	}
}
