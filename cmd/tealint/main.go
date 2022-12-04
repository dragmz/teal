package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dragmz/teal"
	"github.com/pkg/errors"
)

type args struct {
	Path  string
	Stdin bool
}

func run(a args) error {
	if a.Stdin {
		bs, err := io.ReadAll(os.Stdin)
		if err != nil {
			return errors.Wrap(err, "failed to read from stdin")
		}
		p, err := teal.Parse(string(bs))
		if err != nil {
			return errors.Wrap(err, "failed to parse TEAL program")
		}

		l := teal.Compile(p)

		errs := l.Lint()
		for _, err := range errs {
			fmt.Printf("%d: %s\n", err.Line(), err)
		}
	} else {
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
	}

	return nil
}

func main() {
	var a args

	flag.StringVar(&a.Path, "path", "", "source TEAL file path")
	flag.BoolVar(&a.Stdin, "stdin", false, "read from stdin; cannot be used with path")
	flag.Parse()

	if a.Path == "" && !a.Stdin {
		flag.Usage()
		return
	}

	if a.Path != "" && a.Stdin {
		flag.Usage()
		return
	}

	err := run(a)
	if err != nil {
		panic(err)
	}
}
