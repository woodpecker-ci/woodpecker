# go-yaml2json

[![Tests](https://ci.codeberg.org/api/badges/6543/go-yaml2json/status.svg)](https://ci.codeberg.org/6543/go-yaml2json)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/codeberg.org/6543/go-yaml2json?status.svg)](https://godoc.org/codeberg.org/6543/go-yaml2json)
[![Go Report Card](https://goreportcard.com/badge/codeberg.org/6543/go-yaml2json)](https://goreportcard.com/report/codeberg.org/6543/go-yaml2json)

<a href="https://codeberg.org/6543/go-yaml2json">
    <img alt="Get it on Codeberg" src="https://codeberg.org/Codeberg/GetItOnCodeberg/media/branch/main/get-it-on-neon-blue.png" height="60">
</a>

golang lib to convert yaml into json

```sh
go get codeberg.org/6543/go-yaml2json
```

```go
yaml2json.Convert(data []byte) ([]byte, error)
yaml2json.StreamConvert(r io.Reader, w io.Writer) error
```

## Example

[<img src="https://go.dev/images/go-logo-white.svg" alt="Go" height=15> **Playground**](https://go.dev/play/p/fBddDCaucNG)

yaml:

```yaml
- name: Jack
  job: Butcher
- name: Jill
  job: Cook
  obj:
    empty: false
    data: |
      some data 123
      with new line      
```

will become json:
```json
[
  {
    "job": "Butcher",
    "name": "Jack"
  },
  {
    "job": "Cook",
    "name": "Jill",
    "obj": {
      "data": "some data 123\nwith new line\n",
      "empty": false
    }
  }
]
```
