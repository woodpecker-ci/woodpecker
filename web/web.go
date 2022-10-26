package web

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
)

// go:embed dist/*
var webFiles embed.FS

func HTTPFS() http.FileSystem {
	httpFS, err := fs.Sub(webFiles, "dist")
	if err != nil {
		panic(err)
	}

	return http.FS(httpFS)
}

func Lookup(path string) (buf []byte, err error) {
	file, err := HTTPFS().Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf, err = io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func MustLookup(path string) (buf []byte) {
	buf, err := Lookup(path)
	if err != nil {
		panic(err)
	}

	return buf
}
