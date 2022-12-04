package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dragmz/teal"
	"github.com/pkg/errors"
)

type args struct {
	Path string
}

func run(a args) error {
	fi, err := os.Stat(a.Path)
	if err != nil {
		return errors.Wrap(err, "failed to get TEAL file info")
	}

	var paths []string

	if fi.IsDir() {
		infos, err := ioutil.ReadDir(a.Path)
		if err != nil {
			return errors.Wrap(err, "failed to read source directory")
		}

		for _, info := range infos {
			if info.IsDir() {
				continue
			}

			paths = append(paths, filepath.Join(a.Path, info.Name()))
		}
	} else {
		paths = append(paths, a.Path)
	}

	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "failed to read TEAL file")
		}

		p, err := teal.Parse(string(b))
		if err != nil {
			return errors.Wrap(err, "failed to parse TEAL program")
		}

		l := teal.Compile(p)

		errs := l.Lint()
		for _, err := range errs {
			fmt.Printf("%s:%d: %s\n", path, err.Line(), err)
		}
	}

	return nil
}

func main() {
	var a args

	flag.StringVar(&a.Path, "path", "", "source TEAL file path")
	flag.Parse()

	if a.Path == "" {
		flag.Usage()
		return
	}

	err := run(a)
	if err != nil {
		panic(err)
	}
}
