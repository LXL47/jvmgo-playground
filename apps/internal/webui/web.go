package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var content embed.FS

func Handler() http.Handler {
	dist, err := fs.Sub(content, "dist")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(dist))
}
