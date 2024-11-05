# Custom backends

If none of our backends fits your usecase, you can write your own.

Therefore, implement the interface `"go.woodpecker-ci.org/woodpecker/woodpecker/v2/pipeline/backend/types".Backend` and
build a custom agent using your backend with this `main.go`:

```go
package main

import (
  "go.woodpecker-ci.org/woodpecker/v2/cmd/agent/core"
  backendTypes "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

func main() {
  core.RunAgent([]backendTypes.Backend{
    yourBackend,
  })
}
```

It is also possible to use multiple backends, you can select with [`WOODPECKER_BACKEND`](../15-agent-config.md#woodpecker_backend) between them.
