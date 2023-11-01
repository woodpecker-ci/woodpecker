# Creating addons

Addons are written in Go.

## Writing your code

A plugin consists of two variables/functions in Go.

1. The `Type` variable. Specifies the type of the plugin and must be directly accessed from `shared/addons/types/types.go`.
2. The `Addon` function which is the main point of your plugin.
   This function takes two arguments:
    1. The zerolog logger you should use to log errors, warnings etc.
    2. A slice of strings with the environment variables used as configuration.
   
   The function returns two values:
    1. The actual plugin. For type reference see [table below](#return-types).
    2. An error. If this error is not `nil`, Woodpecker exits.

Directly import Woodpecker's Go package (`github.com/woodpecker-ci/woodpecker`) and use the interfaces and types defined there.

### Return types

| Plugin type | Return type |
| --- | --- |
| `Forge` | `"github.com/woodpecker-ci/woodpecker/server/forge".Forge` |
| `Engine` | `"github.com/woodpecker-ci/woodpecker/pipeline/backend/types".Engine` |

## Compiling

After you wrote your addon code, compile your plugin:

```sh
go build -buildmode plugin
```

The output file is your plugin which is now ready to use.

## Restrictions

Plugins must directly directly depend on Woodpecker's core (`github.com/woodpecker-ci/woodpecker`).
The plugin must have been built with **excatly the same code** as the Woodpecker instance you'd like to use it on. This means: If you build your plugin with a specific commit from Woodpecker `next`, you can likely only use it with the Woodpecker version compiled from this commit.
Also, if you change something inside of Woodpecker without commiting, it might fail because you need to recompile your plugin with this code.

:::info
It is recommended to at least support the latest released version of Woodpecker.
:::

### Compile for different versions

As long as there were no changes to Woodpecker's interfaces or they are backwards-compatible, you can easily compile the addon for multiple version by changing the version of `github.com/woodpecker-ci/woodpecker` using `go get` before compiling.

## Example structure

```go
package main

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	addon_types "github.com/woodpecker-ci/woodpecker/shared/addon/types"
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
