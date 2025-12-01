# Packaging

If you repackage it, we encourage to build from source, witch requires internet connection.

We also for offline builds provide a tarball with all dependencies vendored and webui prebuild
on the [release page](https://github.com/woodpecker-ci/woodpecker/releases).

## Distribute WebUI on own directory.

If you want to not embed the webui into the binary, you can compile a custom webui root path into the binary.

Add `external_web` into the tags and use `-X go.woodpecker-ci.org/woodpecker/v3/web.webUIRoot=/some/path` build flag to set custom path.

example:

```sh
go build -tags 'external_web' -ldflags '-s -w -extldflags "-static" -X go.woodpecker-ci.org/woodpecker/v3/version.Version=3.12.0 -X go.woodpecker-ci.org/woodpecker/v3/web.webUIRoot=/nix/store/maaajlp8h5gy9zyjgfhaipzj07qnnmrl-woodpecker-webui-3.12.0' -o dist/woodpecker-server go.woodpecker-ci.org/woodpecker/v3/cmd/server
```
