# Packaging

If you repackage it, we encourage to build from source, which requires internet connection.

For offline builds, we also offer a tarball with all vendored dependencies and a pre-built web UI
on the [release page](https://github.com/woodpecker-ci/woodpecker/releases).

## Distribute web UI in own directory

If you do not want to embed the web UI in the binary, you can compile a custom root path for the web UI into the binary.

Add `external_web` to the tags and use the build flag `-X go.woodpecker-ci.org/woodpecker/v3/web.webUIRoot=/some/path` to set a custom path.

Example: <!-- cspell:ignore webui -->

```sh
go build -tags 'external_web' -ldflags '-s -w -extldflags "-static" -X go.woodpecker-ci.org/woodpecker/v3/version.Version=3.12.0 -X go.woodpecker-ci.org/woodpecker/v3/web.webUIRoot=/nix/store/maaajlp8h5gy9zyjgfhaipzj07qnnmrl-woodpecker-WebUI-3.12.0' -o dist/woodpecker-server go.woodpecker-ci.org/woodpecker/v3/cmd/server
```
