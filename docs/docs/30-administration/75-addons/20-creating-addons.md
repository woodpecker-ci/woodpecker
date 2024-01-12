# Creating addons

Addons are written in Go.

## Writing your code

An addon consists of two variables/functions in Go.

1. The `Type` variable. Specifies the type of the addon and must be directly accessed from `shared/addons/types/types.go`.
2. The `Addon` function which is the main point of your addon.
   This function takes the `zerolog` logger you should use to log errors, warnings, etc. as argument.

   It returns two values:

   1. The actual addon. For type reference see [table below](#return-types).
   2. An error. If this error is not `nil`, Woodpecker exits.

Directly import Woodpecker's Go package (`go.woodpecker-ci.org/woodpecker/woodpecker/v2`) and use the interfaces and types defined there.

### Return types

| Addon type           | Return type                                                                      |
| -------------------- | -------------------------------------------------------------------------------- |
| `Forge`              | `"go.woodpecker-ci.org/woodpecker/woodpecker/v2/server/forge".Forge`             |
| `Backend`            | `"go.woodpecker-ci.org/woodpecker/woodpecker/v2/pipeline/backend/types".Backend` |
| `ConfigService`      | `"go.woodpecker-ci.org/woodpecker/v2/server/plugins/config".Extension`           |
| `SecretService`      | `"go.woodpecker-ci.org/woodpecker/v2/server/model".SecretService`                |
| `EnvironmentService` | `"go.woodpecker-ci.org/woodpecker/v2/server/model".EnvironmentService`           |
| `RegistryService`    | `"go.woodpecker-ci.org/woodpecker/v2/server/model".RegistryService`              |

### Using configurations

If you write a plugin for the server (`Forge` and the services), you can access the server config.

Therefore, use the `"go.woodpecker-ci.org/woodpecker/v2/server".Config` variable.

:::warning
The config is not available when your addon is initialized, i.e., the `Addon` function is called.
Only use the config in the interface methods.
:::

## Compiling

After you write your addon code, compile your addon:

```sh
go build -buildmode plugin
```

The output file is your addon that is now ready to be used.

## Restrictions

Addons must directly depend on Woodpecker's core (`go.woodpecker-ci.org/woodpecker/woodpecker/v2`).
The addon must have been built with **exactly the same code** as the Woodpecker instance you'd like to use it on. This means: If you build your addon with a specific commit from Woodpecker `next`, you can likely only use it with the Woodpecker version compiled from this commit.
Also, if you change something inside Woodpecker without committing, it might fail because you need to recompile your addon with this code first.

In addition to this, addons are only supported on Linux, FreeBSD, and macOS.

:::info
It is recommended to at least support the latest version of Woodpecker.
:::

### Compile for different versions

As long as there are no changes to Woodpecker's interfaces,
or they are backwards-compatible, you can compile the addon for multiple versions
by changing the version of `go.woodpecker-ci.org/woodpecker/woodpecker/v2` using `go get` before compiling.

## Logging

The entrypoint receives a `zerolog.Logger` as input. **Do not use any other logging solution.** This logger follows the configuration of the Woodpecker instance and adds a special field `addon` to the log entries which allows users to find out which component is writing the log messages.

## Example structure

```go
package main

import (
  "context"
  "net/http"

  "github.com/rs/zerolog"
  "go.woodpecker-ci.org/woodpecker/v2/server/forge"
  forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
  "go.woodpecker-ci.org/woodpecker/v2/server/model"
  addon_types "go.woodpecker-ci.org/woodpecker/v2/shared/addon/types"
)

var Type = addon_types.TypeForge

func Addon(logger zerolog.Logger) (forge.Forge, error) {
  logger.Info().Msg("hello world from addon")
  return &config{l: logger}, nil
}

type config struct {
  l zerolog.Logger
}

// In this case, `config` must implement `forge.Forge`. You must directly use Woodpecker's packages - see imports above.
```
