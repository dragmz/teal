package main

import (
	"flag"
	"io"
	"net"
	"os"

	"github.com/dragmz/teal/dbg"

	"github.com/pkg/errors"
)

type args struct {
	Debug string

	Addr string
	Net  string
}

func run(a args) (int, error) {
	var r io.Reader
	var w io.Writer

	if a.Addr != "" && a.Net != "" {
		c, err := net.Dial(a.Net, a.Addr)
		if err != nil {
			return -1, errors.Wrap(err, "failed to connect to the client")
		}

		r = c
		w = c
	} else {
		r = os.Stdin
		w = os.Stdout
	}

	var opts []dbg.DbgOption
	if a.Debug != "" {
		f, err := os.Create(a.Debug)
		if err != nil {
			return -2, errors.Wrap(err, "failed to create debug output file")
		}

		opts = append(opts, dbg.WithDebug(f))
	}

	l, err := dbg.New(r, w, opts...)
	if err != nil {
		return -3, errors.Wrap(err, "failed to create dbg")
	}

	return l.Run()
}

func main() {
	var a args

	flag.StringVar(&a.Net, "net", "tcp", "client network")
	flag.StringVar(&a.Addr, "addr", "", "client address")
	flag.StringVar(&a.Debug, "debug", "E:/dbg.txt", "debug file path")

	flag.Parse()

	code, err := run(a)
	if err != nil {
		panic(err)
	}

	os.Exit(code)
}
