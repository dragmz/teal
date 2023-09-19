package main

import (
	"encoding/json"
	"flag"
	"io"
	"net"
	"os"

	"github.com/dragmz/teal/dbg"
	"github.com/dragmz/teal/lsp"

	"github.com/pkg/errors"
)

type lspArgs struct {
	Debug string

	Addr string
	Net  string
}

type dbgArgs struct {
	Debug      string
	Config     string
	Algod      string
	AlgodToken string
}

func runLsp(a lspArgs) (int, error) {
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

	var opts []lsp.LspOption
	if a.Debug != "" {
		f, err := os.Create(a.Debug)
		if err != nil {
			return -2, errors.Wrap(err, "failed to create debug output file")
		}

		opts = append(opts, lsp.WithDebug(f))
	}

	l, err := lsp.New(r, w, opts...)
	if err != nil {
		return -3, errors.Wrap(err, "failed to create lsp")
	}

	return l.Run()
}

func runDbg(a dbgArgs) (int, error) {
	r := os.Stdin
	w := os.Stdout

	var opts []dbg.DbgOption
	if a.Debug != "" {
		f, err := os.Create(a.Debug)
		if err != nil {
			return -2, errors.Wrap(err, "failed to create debug output file")
		}

		opts = append(opts, dbg.WithDebug(f))
	}

	if a.Config != "" {
		bs, err := os.ReadFile(a.Config)
		if err != nil {
			return -3, errors.Wrap(err, "failed to read config file")
		}

		var cfg dbg.DbgConfig
		if err := json.Unmarshal(bs, &cfg); err != nil {
			return -4, errors.Wrap(err, "failed to unmarshal config file")
		}

		opts = append(opts, dbg.WithConfig(cfg))
	}

	if a.Algod != "" {
		opts = append(opts, dbg.WithAlgod(a.Algod, a.AlgodToken))
	}

	l, err := dbg.New(r, w, opts...)
	if err != nil {
		return -3, errors.Wrap(err, "failed to create dbg")
	}

	return l.Run()
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "dbg" {
		var a dbgArgs

		flag.StringVar(&a.Debug, "debug", "", "debug file path")
		flag.StringVar(&a.Config, "config", "", "config file path")
		flag.StringVar(&a.Algod, "algod", "https://testnet-api.algonode.cloud", "algod endpoint")
		flag.StringVar(&a.AlgodToken, "algod-token", "", "algod token")

		copy(os.Args[1:], os.Args[2:])
		os.Args = os.Args[:len(os.Args)-1]

		flag.Parse()

		code, err := runDbg(a)
		if err != nil {
			panic(err)
		}
		os.Exit(code)
	} else {
		var a lspArgs

		flag.StringVar(&a.Net, "net", "tcp", "client network")
		flag.StringVar(&a.Addr, "addr", "", "client address")
		flag.StringVar(&a.Debug, "debug", "", "debug file path")

		flag.Parse()

		code, err := runLsp(a)
		if err != nil {
			panic(err)
		}

		os.Exit(code)
	}
}
