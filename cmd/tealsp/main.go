package main

import (
	"flag"
	"io"
	"net"
	"os"

	"github.com/dragmz/teal/lsp"

	"github.com/pkg/errors"
)

type args struct {
	Debug string

	Addr string
	Net  string
}

func run(a args) error {
	var r io.Reader
	var w io.Writer

	if a.Addr != "" && a.Net != "" {
		c, err := net.Dial(a.Net, a.Addr)
		if err != nil {
			return errors.Wrap(err, "failed to connect to the client")
		}

		r = c
		w = c
	} else {
		r = os.Stdin
		w = os.Stdout
	}

	var opts []lsp.LspOption
	if a.Debug != "" {
		f, err := os.Create(a.Debug)
		if err != nil {
			return errors.Wrap(err, "failed to create debug output file")
		}

		opts = append(opts, lsp.WithDebug(f))
	}

	l, err := lsp.New(r, w, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create lsp")
	}

	err = l.Run()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	var a args

	flag.StringVar(&a.Net, "net", "tcp", "client network")
	flag.StringVar(&a.Addr, "addr", "", "client address")
	flag.StringVar(&a.Debug, "debug", "", "debug file path")

	flag.Parse()

	err := run(a)
	if err != nil {
		panic(err)
	}
}
