package server

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/rs/zerolog/log"
)

func parseReq[T any](r *http.Request) (T, error) {
	// spew.Dump(r.Header.Get("content-type"))

	return parseReqMultipart[T](r)
}

func parseReqMultipart[T any](r *http.Request) (T, error) {
	var zero T

	if err := r.ParseMultipartForm(1024 * 4); err != nil {
		return zero, fmt.Errorf("parse multipart form: %w", err)
	}

	zeroT := reflect.TypeOf(zero)
	zeroPtrV := reflect.ValueOf(&zero)
	zeroElem := zeroPtrV.Elem()

	type fieldt struct {
		rf         reflect.StructField
		parser     func(any) (any, error)
		validators []func(any) error
	}

	fields := make(map[string]fieldt, zeroT.NumField())

	for i := 0; i < zeroT.NumField(); i++ {
		rf := zeroT.Field(i)

		fields[strings.ToLower(rf.Name)] = fieldt{
			rf:         rf,
			parser:     getFieldParser(rf.Type),
			validators: parseValidators(rf.Tag.Get("validate")),
		}
	}

	for name, vals := range r.MultipartForm.Value {
		if vals[0] == "" {
			continue
		}

		f, ok := fields[strings.ToLower(name)]
		if !ok {
			log.Debug().Str("field", name).Msg("skip missing field")
			continue
		}

		val, err := f.parser(vals[0])
		if err != nil {
			return zero, err
		}

		zeroElem.FieldByIndex(f.rf.Index).Set(reflect.ValueOf(val))
	}

	for name, f := range fields {
		for _, v := range f.validators {
			if err := v(zeroElem.FieldByIndex(f.rf.Index).Interface()); err != nil {
				return zero, internal.NewPubErr(fmt.Errorf("%w %q: %s", internal.ErrInvalid, name, err), nil)
			}
		}
	}

	for name, vals := range r.MultipartForm.File {
		f, ok := fields[strings.ToLower(name)]
		if !ok {
			log.Debug().Str("field", name).Msg("skip missing field file")
			continue
		}

		zeroElem.FieldByIndex(f.rf.Index).Set(reflect.ValueOf(vals[0]))
	}

	return zero, nil
}

func getFieldParser(t reflect.Type) func(val any) (any, error) {
	switch t.Kind() {
	case reflect.Int64:
		return func(val any) (any, error) {
			out, err := strconv.ParseInt(val.(string), 10, 64)
			if err != nil {
				return nil, internal.NewPubErr(internal.ErrInvalid, err)
			}

			return out, nil
		}
	case reflect.Float64:
		return func(val any) (any, error) {
			out, err := strconv.ParseFloat(val.(string), 64)
			if err != nil {
				return nil, internal.NewPubErr(internal.ErrInvalid, err)
			}

			return out, nil
		}
		// case reflect.Ptr:
		// 	switch {
		// 	case t.AssignableTo(reflect.TypeOf(&multipart.FileHeader{})):
		// 		return func(val any) (any, error) {
		// 			return val, nil
		// 		}
		// 	}
		// case reflect.Struct:
		// 	switch {
		// 	}
	}

	return func(val any) (any, error) {
		return val, nil
	}
}

func parseValidators(tag string) (out []func(any) error) {
	if len(tag) == 0 {
		return
	}

	for _, part := range strings.Split(tag, ";") {
		switch part {
		case "required":
			out = append(out, func(a any) error {
				v := reflect.ValueOf(a)
				if v.IsZero() {
					return errors.New("required")
				}

				return nil
			})
		default:
			panic(fmt.Sprintf("unknown validator: %q", part))
		}
	}

	return
}
