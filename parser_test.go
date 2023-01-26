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

func testParser(t *testing.T, dir string, clean bool) {
	fs, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range fs {
		if path.Ext(f.Name()) != ".teal" {
			continue
		}

		bs, err := os.ReadFile(path.Join(dir, f.Name()))
		if !assert.NoError(t, err) {
			return
		}

		res := Process(string(bs))
		if clean {
			assert.Empty(t, res.Diagnostics)
		} else {
			assert.NotEmpty(t, res.Diagnostics)
		}
	}
}

func TestParserOk(t *testing.T) {
	testParser(t, filepath.Join("examples", "ok"), true)
}

func TestParserErr(t *testing.T) {
	testParser(t, filepath.Join("examples", "err"), false)
}
