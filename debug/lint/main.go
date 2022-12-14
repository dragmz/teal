package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dragmz/teal"
)

func main() {
	var path string

	flag.StringVar(&path, "path", "", "path to teal file")
	flag.Parse()

	s, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	res := teal.Process(string(s))
	for _, d := range res.Diagnostics {
		fmt.Printf("%d:%d-%d:%s %s\n", d.Line(), d.Begin(), d.End(), d.Severity(), d)
	}
}
