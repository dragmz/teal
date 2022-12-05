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

	p, errs := teal.Parse(string(s))
	for _, err := range errs {
		fmt.Printf("%d: %s\n", err.Line(), err)
	}

	if len(errs) == 0 {
		fmt.Println(p)
	}
}
