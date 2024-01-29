package ksqlite

import (
	"errors"
	"fmt"

	"github.com/eatonphil/gosqlite"
)

type Conn struct {
	conn  *gosqlite.Conn
	stmts map[string]*gosqlite.Stmt
}

func NewConn(conn *gosqlite.Conn) *Conn {
	return &Conn{
		conn:  conn,
		stmts: map[string]*gosqlite.Stmt{},
	}
}

func (c *Conn) Close() error {
	for _, stmt := range c.stmts {
		if err := stmt.Close(); err != nil {
			return fmt.Errorf("close stmt: %w", err)
		}
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("close sqlite conn: %w", err)
	}
	return nil
}

func (c *Conn) Prepare(sql string, args ...interface{}) (stmt *gosqlite.Stmt, outErr error) {
	defer func() {
		if outErr != nil {
			delete(c.stmts, sql)

			if stmt == nil {
				return
			}

			if err := stmt.Close(); err != nil {
				outErr = errors.Join(outErr, fmt.Errorf("close stmt after bad bind: %w", err))
			}
			return
		}

		if err := stmt.Reset(); err != nil {
			outErr = fmt.Errorf("reset stmt: %w", err)
			return
		}

		if err := stmt.Bind(args...); err != nil {
			outErr = fmt.Errorf("bind args to stmt: %w", err)
			return
		}
	}()

	if stmt, ok := c.stmts[sql]; ok {
		return stmt, nil
	}

	stmt, err := c.conn.Prepare(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("prepare stmt: %w", err)
	}

	c.stmts[sql] = stmt
	return stmt, nil
}
