package server

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/davecgh/go-spew/spew"
)

func parseReq[T any](r *http.Request) (T, error) {
	// spew.Dump(r.Header.Get("content-type"))

	return parseReqMultipart[T](r)
}

func parseReqMultipart[T any](r *http.Request) (T, error) {
	var zero T

	rt := reflect.TypeOf(zero)
	spew.Dump(rt)

	// todo: cache this
	fields := make(map[string]func(string) (any, error), rt.NumField())

	for i := 0; i < rt.NumField(); i++ {
		rf := rt.Field(i)
		spew.Dump(rf)

		switch rf.Type.Kind() {
		case reflect.String:
			fields[strings.ToLower(rf.Name)] = func(val string) (any, error) {
				return val, nil
			}
		case reflect.Int64:
			fields[strings.ToLower(rf.Name)] = func(val string) (any, error) {
				out, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return nil, internal.NewSErr(errors.New("invalid kcal format"), err)
				}

				return out, nil
			}
		case reflect.Float64:
			fields[strings.ToLower(rf.Name)] = func(val string) (any, error) {
				out, err := strconv.ParseFloat(val, 64)
				if err != nil {
					return nil, internal.NewSErr(errors.New("invalid kcal format"), err)
				}

				return out, nil
			}
		case reflect.Struct:
			switch rf.Type.PkgPath() {
			case "github.com/GerardRodes/kcalc/internal.Image":
				switch rf.Type.Name() {
				case "Image":
					fields[strings.ToLower(rf.Name)] = func(name string) (any, error) {
						spew.Dump("internal.Image", name)
						return nil, nil
					}
				}
			}
		default:
			panic(fmt.Sprintf("unsupported field kind %q", rf.Type.Kind()))
		}
	}

	return zero, nil
}
