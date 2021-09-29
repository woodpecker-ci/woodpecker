package web

import (
	"embed"
	"io/fs"
	"io/ioutil"
	"net/http"
)

//go:embed dist/*
var webFiles embed.FS

func HttpFS() http.FileSystem {
	httpFS, err := fs.Sub(webFiles, "dist")
	if err != nil {
		panic(err)
	}

	return http.FS(httpFS)
}

func Lookup(path string) (buf []byte, err error) {
	file, err := HttpFS().Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf, err = ioutil.ReadAll(file)
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
