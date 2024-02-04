package server

import (
	"embed"
	"html/template"
	"path/filepath"
	"reflect"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/rs/zerolog/log"
)

//go:embed templates/*/*.tmpl
var gohtml embed.FS

func refVal(v any) reflect.Value {
	if val, ok := v.(reflect.Value); ok {
		return val
	}

	return reflect.ValueOf(v)
}

var tmpl = template.New("").Funcs(template.FuncMap{
	"last": func(v any) any {
		val := refVal(v)
		return val.Index(val.Len() - 1)
	},
	"field": func(f string, v any) any {
		return refVal(v).FieldByName(f)
	},
	"last_iter": func(v any, i int64) bool {
		return refVal(v).Len() == int(i+1)
	},
})

func init() {
	tmpl = template.Must(tmpl.ParseFS(gohtml, "templates/fragments/*.tmpl"))

	for _, f := range internal.Must(gohtml.ReadDir("templates/views")) {
		name := f.Name()[:len(f.Name())-len(filepath.Ext(f.Name()))]
		tmpl = template.Must(tmpl.New(name).ParseFS(gohtml, filepath.Join("templates/views", f.Name())))
		log.Debug().Str("name", name).Msg("parsed template")
	}
}

func newData() map[any]any {
	return map[any]any{
		"langByID":   internal.LangByID,
		"sourceByID": internal.SourceByID,
	}
}
