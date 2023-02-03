package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

const LATEST = 9

func makeUrl(version int) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/barnjamin/go-algorand/extend-stack-type/data/transactions/logic/langspec_v%d.json", version)
}

type args struct {
	Path string
}

func run(a args) error {
	os.MkdirAll(a.Path, os.ModePerm)

	for i := 1; i <= LATEST; i++ {
		url := makeUrl(i)

		resp, err := http.Get(url)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return errors.New(resp.Status)
		}

		p := path.Join(a.Path, fmt.Sprintf("langspec_v%d.json", i))
		f, err := os.Create(p)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var a args
	flag.StringVar(&a.Path, "path", "", "target path")
	flag.Parse()

	err := run(a)
	if err != nil {
		panic(err)
	}
}
