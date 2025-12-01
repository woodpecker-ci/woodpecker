# Packaging

If you repackage it, we encourage to build from source, which requires internet connection.

For offline builds we also provide a tarball with all dependencies vendored and web UI prebuilt
on the [release page](https://github.com/woodpecker-ci/woodpecker/releases).

## Distribute web UI in own directory

If you do not want to embed the web UI into the binary, you can compile a custom web UI root path into the binary.

Add `external_web` into the tags and use `-X go.woodpecker-ci.org/woodpecker/v3/web.webUIRoot=/some/path` build flag to set custom path.

example: <!-- cspell:ignore webui -->

```sh
go build -tags 'external_web' -ldflags '-s -w -extldflags "-static" -X go.woodpecker-ci.org/woodpecker/v3/version.Version=3.12.0 -X go.woodpecker-ci.org/woodpecker/v3/web.webUIRoot=/nix/store/maaajlp8h5gy9zyjgfhaipzj07qnnmrl-woodpecker-WebUI-3.12.0' -o dist/woodpecker-server go.woodpecker-ci.org/woodpecker/v3/cmd/server
```
