# Creating addons

Addons are written in Go.

## Writing your code

An addon consists of two variables/functions in Go.

1. The `Type` variable. Specifies the type of the addon and must be directly accessed from `shared/addons/types/types.go`.
2. The `Addon` function which is the main point of your addon.
   This function takes two arguments:

   1. The zerolog logger you should use to log errors, warnings etc.
   2. A slice of strings with the environment variables used as configuration.

   It returns two values:

   1. The actual addon. For type reference see [table below](#return-types).
   2. An error. If this error is not `nil`, Woodpecker exits.

Directly import Woodpecker's Go package (`go.woodpecker-ci.org/woodpecker/woodpecker`) and use the interfaces and types defined there.

### Return types

| Addon type | Return type                                                                  |
| ---------- | ---------------------------------------------------------------------------- |
| `Forge`    | `"go.woodpecker-ci.org/woodpecker/woodpecker/server/forge".Forge`            |
| `Engine`   | `"go.woodpecker-ci.org/woodpecker/woodpecker/pipeline/backend/types".Engine` |

## Compiling

After you wrote your addon code, compile your addon:

```sh
go build -buildmode plugin
```

The output file is your addon which is now ready to use.

## Restrictions

Addons must directly directly depend on Woodpecker's core (`go.woodpecker-ci.org/woodpecker/woodpecker`).
The addon must have been built with **excatly the same code** as the Woodpecker instance you'd like to use it on. This means: If you build your addon with a specific commit from Woodpecker `next`, you can likely only use it with the Woodpecker version compiled from this commit.
Also, if you change something inside of Woodpecker without commiting, it might fail because you need to recompile your addon with this code.

In addition to this, addons are only supported on Linux, FreeBSD and macOS.

:::info
It is recommended to at least support the latest released version of Woodpecker.
:::

### Compile for different versions

As long as there were no changes to Woodpecker's interfaces or they are backwards-compatible, you can easily compile the addon for multiple version by changing the version of `go.woodpecker-ci.org/woodpecker/woodpecker` using `go get` before compiling.

## Example structure

```go
package main

import (
  "context"
  "net/http"

  "github.com/rs/zerolog"
  "go.woodpecker-ci.org/woodpecker/woodpecker/server/forge"
  forge_types "go.woodpecker-ci.org/woodpecker/woodpecker/server/forge/types"
  "go.woodpecker-ci.org/woodpecker/woodpecker/server/model"
  addon_types "go.woodpecker-ci.org/woodpecker/woodpecker/shared/addon/types"
)

var Type = addon_types.TypeForge

func Addon(logger zerolog.Logger, env []string) (forge.Forge, error) {
  logger.Info().Msg("hello world from addon")
  return &config{l: logger}, nil
}

type config struct {
  l zerolog.Logger
}

// ... in this case, `config` must implement `forge.Forge`. You must directly use Woodpecker's packages - see imports above.
```
