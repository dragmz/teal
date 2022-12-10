# teal

Go packages for Algorand TEAL

##  tealsp

TEAL LSP Server for LSP-compatible editors, currently used in:

- vscode-teal - Visual Studio Code extension: https://marketplace.visualstudio.com/items?itemName=DragMZ.teal

## types

```go
package main

import (
    "fmt"

    "github.com/dragmz/teal"
)

func main() {
    l1 := teal.Label("l1")

    prog := teal.Program{
        l1,
        teal.B(l1),
    }

    fmt.Println(prog)
}
```
