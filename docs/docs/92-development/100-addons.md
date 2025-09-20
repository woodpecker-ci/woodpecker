# Addons

The Woodpecker server supports addons for forges and the log store.

:::warning
Addons are still experimental. Their implementation can change and break at any time.
:::

:::danger
You must trust the author of the addon you are using. They may have access to authentication codes and other potentially sensitive information.
:::

## Bug reports

If you experience bugs, please check which component has the issue. If it's the addon, **do not raise an issue in the main repository**, but rather use the separate addon repositories. To check which component is responsible for the bug, look at the logs. Logs from addons are marked with a special field `addon` containing their addon file name.

## Creating addons

Addons use RPC to communicate to the server and are implemented using the [`go-plugin` library](https://github.com/hashicorp/go-plugin).

### Writing your code

This example will use the Go language.

Directly import Woodpecker's Go packages (`go.woodpecker-ci.org/woodpecker/v3`) and use the interfaces and types defined there.

In the `main` function, just call the `Serve` in the corresponding addon package with the service as argument (see [below](#addon-types)).
This will take care of connecting the addon forge to the server.

:::note
It is not possible to access global variables from Woodpecker, for example the server configuration. You must therefore parse the environment variables in your addon. The reason for this is that the addon runs in a completely separate process.
:::

### Example structure

This is an example for a forge addon.

```go
package main

import (
  "context"
  "net/http"

  "go.woodpecker-ci.org/woodpecker/v3/server/forge/addon"
  forgeTypes "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
  "go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func main() {
  addon.Serve(config{})
}

type config struct {
}

// `config` must implement `"go.woodpecker-ci.org/woodpecker/v3/server/forge".Forge`. You must directly use Woodpecker's packages - see imports above.
```

### Addon types

| Type      | Addon package                                                 | Service interface                                                 |
| --------- | ------------------------------------------------------------- | ----------------------------------------------------------------- |
| Forge     | `go.woodpecker-ci.org/woodpecker/v3/server/forge/addon`       | `"go.woodpecker-ci.org/woodpecker/v3/server/forge".Forge`         |
| Log store | `go.woodpecker-ci.org/woodpecker/v3/server/service/log/addon` | `"go.woodpecker-ci.org/woodpecker/v3/server/service/log".Service` |
