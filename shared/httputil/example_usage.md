# User-Agent RoundTripper Usage Examples

This document demonstrates how to use the `UserAgentRoundTripper` to add custom Woodpecker User-Agent headers to HTTP requests.

## Overview

The `UserAgentRoundTripper` is an `http.RoundTripper` that automatically adds a Woodpecker-specific User-Agent header to all outgoing HTTP requests. This helps identify Woodpecker requests in web access logs.

The User-Agent format is: `Woodpecker/<version> (<component>)`

For example: `Woodpecker/v3.0.0 (cli)` or `Woodpecker/dev (forge-github)`

## Basic Usage

### 1. Wrapping an Existing HTTP Client

The simplest way is to use `WrapClient`:

```go
import (
    "net/http"
    "go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

// Wrap an existing client
client := &http.Client{}
wrappedClient := httputil.WrapClient(client, "my-component")

// Now all requests will include the custom User-Agent
resp, err := wrappedClient.Get("https://api.example.com")
```

### 2. Creating a New RoundTripper

For more control, create a RoundTripper directly:

```go
import (
    "net/http"
    "go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

// Create a custom transport
baseTransport := &http.Transport{
    MaxIdleConns: 100,
}

// Wrap it with User-Agent support
userAgentTransport := httputil.NewUserAgentRoundTripper(baseTransport, "my-service")

// Use in an HTTP client
client := &http.Client{
    Transport: userAgentTransport,
}
```

## Advanced Usage Examples

### Example 1: With OAuth2

When using OAuth2, wrap the base transport:

```go
import (
    "context"
    "net/http"
    "golang.org/x/oauth2"
    "go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

ctx := context.Background()
ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "token"})
tc := oauth2.NewClient(ctx, ts)

// Get the oauth2 transport
tp, _ := tc.Transport.(*oauth2.Transport)

// Wrap the base transport with User-Agent
tp.Base = httputil.NewUserAgentRoundTripper(http.DefaultTransport, "oauth-client")
```

### Example 2: With Custom TLS Configuration

```go
import (
    "crypto/tls"
    "net/http"
    "go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

// Create custom transport with TLS config
baseTransport := &http.Transport{
    TLSClientConfig: &tls.Config{
        InsecureSkipVerify: true, // Don't do this in production!
    },
    Proxy: http.ProxyFromEnvironment,
}

// Wrap with User-Agent
transport := httputil.NewUserAgentRoundTripper(baseTransport, "custom-tls-client")

client := &http.Client{
    Transport: transport,
}
```

### Example 3: Chaining Multiple RoundTrippers

RoundTrippers can be chained for multiple behaviors:

```go
import (
    "net/http"
    "go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

// Example: Custom logging RoundTripper
type LoggingRoundTripper struct {
    base http.RoundTripper
}

func (rt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    log.Printf("Request: %s %s", req.Method, req.URL)
    return rt.base.RoundTrip(req)
}

// Chain: Base -> UserAgent -> Logging
baseTransport := http.DefaultTransport
userAgentRT := httputil.NewUserAgentRoundTripper(baseTransport, "chained-client")
loggingRT := &LoggingRoundTripper{base: userAgentRT}

client := &http.Client{
    Transport: loggingRT,
}
```

### Example 4: Preserving Existing User-Agent

The RoundTripper only sets the User-Agent if it's not already present:

```go
import (
    "net/http"
    "go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

client := httputil.WrapClient(nil, "test")

// This request will use the Woodpecker User-Agent
req1, _ := http.NewRequest("GET", "https://api.example.com", nil)
client.Do(req1) // User-Agent: Woodpecker/vX.X.X (test)

// This request will keep its custom User-Agent
req2, _ := http.NewRequest("GET", "https://api.example.com", nil)
req2.Header.Set("User-Agent", "CustomAgent/1.0")
client.Do(req2) // User-Agent: CustomAgent/1.0
```

## Component Naming Conventions

When choosing a component name, follow these conventions:

- **CLI tools**: `"cli"`
- **Agent**: `"agent"`
- **Server**: `"server"` or `"server-<purpose>"` (e.g., `"server-extensions"`)
- **Forge integrations**: `"forge-<name>"` (e.g., `"forge-github"`, `"forge-gitlab"`)
- **Go client library**: `"go-client"`
- **Backend**: `"backend-<type>"` (e.g., `"backend-docker"`)

## Testing

To verify the User-Agent is set correctly:

```go
import (
    "net/http"
    "net/http/httptest"
    "testing"
    "go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

func TestUserAgent(t *testing.T) {
    // Create test server
    var capturedUA string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        capturedUA = r.Header.Get("User-Agent")
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    // Make request with wrapped client
    client := httputil.WrapClient(nil, "test")
    req, _ := http.NewRequest("GET", server.URL, nil)
    client.Do(req)

    // Verify User-Agent
    if !strings.Contains(capturedUA, "Woodpecker/") {
        t.Errorf("Expected Woodpecker User-Agent, got: %s", capturedUA)
    }
}
```

## Migration Guide

If you have existing code using `http.DefaultClient` or custom clients:

**Before:**
```go
client := http.DefaultClient
resp, err := client.Get("https://api.example.com")
```

**After:**
```go
import "go.woodpecker-ci.org/woodpecker/v3/shared/httputil"

client := httputil.WrapClient(http.DefaultClient, "my-component")
resp, err := client.Get("https://api.example.com")
```

**Before (with custom transport):**
```go
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: tlsConfig,
    },
}
```

**After:**
```go
import "go.woodpecker-ci.org/woodpecker/v3/shared/httputil"

baseTransport := &http.Transport{
    TLSClientConfig: tlsConfig,
}
client := &http.Client{
    Transport: httputil.NewUserAgentRoundTripper(baseTransport, "my-component"),
}
```

## Notes

- The User-Agent includes the Woodpecker version from `version.String()`
- If version is not set (development builds), it defaults to `"dev"` but the value will be fetched from the commit tag which is not a problem.
- The RoundTripper clones requests before modifying them to avoid side effects
- Existing User-Agent headers are preserved and not overwritten

