package server

import (
	"embed"
	"html/template"
	"math/rand"
	"path/filepath"
	"reflect"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
)

//go:embed templates/**/*.tmpl
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
	"dump": func(v any) string {
		return spew.Sdump(v)
	},
	"rand_uint64": func() uint64 {
		return rand.Uint64()
	},
	"get_locale": func(l map[int64]internal.Locale, langID int64) string {
		if v, ok := l[langID]; ok && v.Value != "" {
			return v.Value
		}

		if v, ok := l[internal.LangsID["en"]]; ok && v.Value != "" {
			return v.Value
		}
		if v, ok := l[internal.LangsID["es"]]; ok && v.Value != "" {
			return v.Value
		}

		for _, v := range l {
			if v.Value != "" {
				return v.Value
			}
		}

		return ""
	},
	"map": func(m map[any]any, kv ...any) map[any]any {
		if m == nil {
			m = map[any]any{}
		}

		for i := 0; i < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}

		return m
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

func newData(data map[any]any) map[any]any {
	if data == nil {
		data = map[any]any{}
	}
	data["langID"] = internal.LangsID["en"]
	data["langByID"] = internal.LangByID
	data["sourceByID"] = internal.SourceByID
	data["isProd"] = internal.IsProd
	return data
}
