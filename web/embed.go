// Package web embeds the web assets (templates and static files) into the binary.
package web

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed templates/*.html templates/cv/*.tex
var templateRaw embed.FS

//go:embed static/*
var staticRaw embed.FS

// TemplateFS is the sub-filesystem rooted at the web/templates directory.
var TemplateFS fs.FS = templateRaw

// StaticFS is the sub-filesystem rooted at web/static.
var StaticFS fs.FS

func init() {
	sub, err := fs.Sub(staticRaw, "static")
	if err != nil {
		log.Fatalf("web: sub static: %v", err)
	}
	StaticFS = sub
}
