package ksqlite

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"

	"github.com/GerardRodes/kcalc/internal"
)

// todo: make KV* generic

func KVSet(k string, v any) error {
	var val any

	switch vT := v.(type) {
	case uint, uint8, uint16, uint32, uint64:
		var data [8]byte
		binary.LittleEndian.PutUint64(data[:], reflect.ValueOf(vT).Uint())
		val = data[:]
	default:
		return fmt.Errorf("unsupported type: %T", vT)
	}

	return WExec(`
		insert into kv (k, v)
		values (?, ?)
		on conflict do update
		set v = excluded.v;
		`, k, val)
}

func KVGet[T any](k string) (t T, outErr error) {
	var zero T

	c, unlock := RConn()
	defer unlock()
	stmt, err := c.Prepare("select v from kv where k = ?", k)
	if err != nil {
		return zero, fmt.Errorf("prepare: %w", internal.NewErrWithStackTrace(err))
	}
	defer func() {
		if err := stmt.Reset(); err != nil {
			outErr = errors.Join(outErr, fmt.Errorf("stmt reset: %w", err))
		}
	}()

	hasRow, err := stmt.Step()
	if err != nil {
		return zero, fmt.Errorf("step stmt: %w", internal.NewErrWithStackTrace(err))
	}
	if !hasRow {
		return zero, internal.NewErrWithStackTrace(internal.ErrNotFound)
	}

	var val any

	switch zeroT := any(zero).(type) {
	case uint, uint8, uint16, uint32, uint64:
		val = [8]byte{}
	default:
		return zero, fmt.Errorf("unsupported type: %T", zeroT)
	}

	if err = stmt.Scan(&val); err != nil {
		return zero, fmt.Errorf("scan stmt: %w", internal.NewErrWithStackTrace(err))
	}

	switch zeroT := any(zero).(type) {
	case uint, uint8, uint16, uint32, uint64:
		valC, ok := val.([]byte)
		if !ok {
			return zero, internal.NewErrWithStackTrace(fmt.Errorf("invalid type: %T", val))
		}
		n := binary.LittleEndian.Uint64(valC)
		return any(n).(T), nil
		// return *(*T)(unsafe.Pointer(&n)), nil
	default:
		return zero, fmt.Errorf("unsupported type: %T", zeroT)
	}
}
