package server

import (
	"embed"
	"html/template"
	"path/filepath"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/rs/zerolog/log"
)

//go:embed templates/*/*.tmpl
var gohtml embed.FS

var tmpl = template.New("")

func init() {
	tmpl = template.Must(tmpl.ParseFS(gohtml, "templates/fragments/*.tmpl"))

	for _, f := range internal.Must(gohtml.ReadDir("templates/views")) {
		name := f.Name()[:len(f.Name())-len(filepath.Ext(f.Name()))]
		tmpl = template.Must(tmpl.New(name).ParseFS(gohtml, filepath.Join("templates/views", f.Name())))
		log.Debug().Str("name", name).Msg("parsed template")
	}
}
