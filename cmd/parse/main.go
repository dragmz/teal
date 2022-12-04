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

	p, err := teal.Parse(string(s))
	if err != nil {
		panic(err)
	}

	fmt.Println(p)
}
