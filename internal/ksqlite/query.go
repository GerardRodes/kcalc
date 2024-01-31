package ksqlite

import (
	"fmt"
	"reflect"

	"github.com/GerardRodes/kcalc/internal"
)

func WQueryOne[T any](sql string, args ...any) (T, error) {
	c, unlock := WConn()
	defer unlock()
	return QueryOne[T](c, sql, args...)
}

func RQueryOne[T any](sql string, args ...any) (T, error) {
	c, unlock := RConn()
	defer unlock()
	return QueryOne[T](c, sql, args...)
}

func QueryOne[T any](c *Conn, sql string, args ...any) (T, error) {
	var zero T

	rows, err := Query[T](c, sql, args...)
	if err != nil {
		return zero, fmt.Errorf("query: %w", err)
	}

	if len(rows) == 0 {
		return zero, internal.ErrNotFound
	}

	return rows[0], nil
}

func WQuery[T any](sql string, args ...any) (rows []T, err error) {
	c, unlock := WConn()
	defer unlock()
	return Query[T](c, sql, args...)
}

func RQuery[T any](sql string, args ...any) (rows []T, err error) {
	c, unlock := RConn()
	defer unlock()
	return Query[T](c, sql, args...)
}

func Query[T any](c *Conn, sql string, args ...any) (rows []T, err error) {
	stmt, err := c.Prepare(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}

	var zero T
	var fieldPtrs []any
	isStruct := reflect.TypeOf(zero).Kind() == reflect.Struct
	if isStruct {
		fieldPtrs = make([]any, reflect.ValueOf(zero).NumField())
	} else {
		fieldPtrs = make([]any, 1)
	}

	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, fmt.Errorf("step stmt: %w", err)
		}
		if !hasRow {
			// The query is finished
			break
		}

		var row T

		if isStruct {
			elem := reflect.ValueOf(&row).Elem()
			for i := 0; i < elem.NumField(); i++ {
				fieldPtrs[i] = elem.Field(i).Addr().Interface()
			}
		} else {
			fieldPtrs = []any{&row}
		}

		err = stmt.Scan(fieldPtrs...)
		if err != nil {
			return nil, fmt.Errorf("scan stmt: %w", err)
		}

		rows = append(rows, row)
	}

	return
}

func Exec(sql string, args ...any) error {
	c, unlock := WConn()
	defer unlock()

	stmt, err := c.Prepare(sql, args...)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}

	if err := stmt.Exec(args...); err != nil {
		return fmt.Errorf("exec stmt: %w", err)
	}

	return nil
}
