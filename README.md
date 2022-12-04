# teal

Go types for Algorand TEAL

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
