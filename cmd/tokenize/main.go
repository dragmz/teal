package main

import (
	"flag"
	"fmt"
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

	z := teal.Lexer{Source: bs}
	for z.Scan() {
		t := z.Curr()
		switch t.Type() {
		case teal.TokenEol:
			fmt.Printf("%d:%d:%d: %s\n", t.Line()+1, t.Begin()+1, t.End()+1, t.Type())
		default:
			fmt.Printf("%d:%d:%d: %s = %s\n", t.Line()+1, t.Begin()+1, t.End()+1, t.Type(), t)
		}
	}

	for _, err := range z.Errors() {
		fmt.Println(err)
	}

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
