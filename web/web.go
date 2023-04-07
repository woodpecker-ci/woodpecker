package web

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var webFiles embed.FS

func HTTPFS() (http.FileSystem, error) {
	httpFS, err := fs.Sub(webFiles, "dist")
	if err != nil {
		return nil, err
	}
	return http.FS(httpFS), nil
}

func Lookup(path string) (buf []byte, err error) {
	httpFS, err := HTTPFS()
	if err != nil {
		return nil, err
	}
	file, err := httpFS.Open(path)
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
