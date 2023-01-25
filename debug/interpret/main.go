package main

import (
	"flag"
	"os"

	"github.com/dragmz/teal"
	"github.com/pkg/errors"
)

type args struct {
	Path string
}

func run(a args) error {
	bs, err := os.ReadFile(a.Path)
	if err != nil {
		return errors.Wrap(err, "failed to read source file")
	}

	res := teal.Process(string(bs))

	vm, err := teal.NewVm(res)
	if err != nil {
		return errors.Wrap(err, "failed to create vm")
	}

	vm.Run()
	return nil
}

func main() {
	var a args
	flag.StringVar(&a.Path, "path", "", "source file path")
	flag.Parse()

	err := run(a)
	if err != nil {
		panic(err)
	}
}
