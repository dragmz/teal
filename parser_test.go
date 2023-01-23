package teal

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocs(t *testing.T) {
	i, ok := Ops.Get(OpContext{
		Name:    "txn",
		Version: 9,
	})

	if !ok {
		t.Error("txn not found")
	}

	if i.Name != "txn" {
		t.Error("unexpected name")
	}
}

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
			if !assert.NoError(t, err) {
				return
			}

			res := Process(string(bs))
			if test.Clean {
				assert.Empty(t, res.Diagnostics)
			} else {
				assert.NotEmpty(t, res.Diagnostics)
			}
		}
	}
}
