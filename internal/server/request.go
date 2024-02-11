package server

import (
	"net/http"
	"reflect"
	"strings"

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

	type fieldt struct {
		parser func(string) any
	}
	fields := make(map[string]fieldt, rt.NumField())

	for i := 0; i < rt.NumField(); i++ {
		rf := rt.Field(i)
		spew.Dump(rf)
		switch rf.Type.Kind() {
		case reflect.Int64:
			fields[strings.ToLower(rf.Name)] = fieldt{}
		default:
			spew.Dump(rt)
			panic("unsupported field kind")
		}
	}

	return zero, nil
}
