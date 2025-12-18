---
sidebar_position: 1
---

# Installation

## Requirements

- **Go 1.21 or later** - Pedantigo uses generics (`New[T]()`, `Unmarshal[T]()`)

## Install

```bash
go get github.com/SmrutAI/pedantigo
```

## Verify Installation

Create a simple test:

```go
package main

import (
    "fmt"
    "github.com/smrutai/pedantigo"
)

type User struct {
    Email string `json:"email" pedantigo:"required,email"`
}

func main() {
    user, err := pedantigo.Unmarshal[User]([]byte(`{"email":"test@example.com"}`))
    if err != nil {
        panic(err)
    }
    fmt.Println("Installation verified:", user.Email)
}
```

Run it:
```bash
go run main.go
```

Expected output:
```
Installation verified: test@example.com
```

## IDE Support

:::tip VS Code
Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go) for struct tag autocomplete and validation.
:::

## Next Steps

Ready to build? Head to the [Quick Start guide](./quickstart).
