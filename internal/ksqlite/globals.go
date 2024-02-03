package ksqlite

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/rs/zerolog/log"
)

var (
	wl     sync.Mutex
	w      *Conn
	rls    []*sync.Mutex
	rconns []*Conn
)

func WConn() (*Conn, func()) {
	wl.Lock()

	return w, wl.Unlock
}

func RConn() (*Conn, func()) {
	for i := range rls {
		if rls[i].TryLock() {
			return rconns[i], rls[i].Unlock
		}
	}

	i := rand.Intn(len(rls))
	rls[i].Lock()
	return rconns[i], rls[i].Unlock
}

func InitGlobals(name string, readConns int, create bool) error {
	if err := InitWrite(name, create); err != nil {
		return fmt.Errorf("init write: %w", err)
	}
	log.Debug().Msg("init sqlite write conn")

	if err := RunMigrations(); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Debug().Msg("ran sqlite migrations")

	if err := InitReads(name, readConns); err != nil {
		return fmt.Errorf("init reads: %w", err)
	}
	log.Debug().Msgf("init %d sqlite read conns", readConns)

	var err error
	internal.LangByID, err = ListLangsByID()
	if err != nil {
		return fmt.Errorf("init langs: %w", err)
	}

	internal.SourceByID, err = ListSourcesByID()
	if err != nil {
		return fmt.Errorf("init sources: %w", err)
	}

	return nil
}

func CloseGlobals() error {
	wl.Lock()

	if err := w.Close(); err != nil {
		return fmt.Errorf("close write conn: %w", err)
	}

	for i := range rls {
		rls[i].Lock()
		if err := rconns[i].Close(); err != nil {
			return fmt.Errorf("close read conn: %w", err)
		}
	}
	return nil
}

func InitWrite(name string, create bool) error {
	flags := SQLITE_OPEN_READWRITE
	if create {
		flags = flags | SQLITE_OPEN_CREATE
	}
	conn, err := Open(name, flags)
	if err != nil {
		return fmt.Errorf("open write conn: %w", err)
	}

	w = NewConn(conn)

	return nil
}

func InitReads(name string, readConns int) error {
	rls = make([]*sync.Mutex, readConns)
	rconns = make([]*Conn, readConns)

	for i := 0; i < readConns; i++ {
		conn, err := Open(name, SQLITE_OPEN_READONLY)
		if err != nil {
			return fmt.Errorf("open read conn: %w", err)
		}

		rls[i] = &sync.Mutex{}
		rconns[i] = NewConn(conn)
	}

	return nil
}

func Optimize() error {
	log.Debug().Msg("optimizing db")
	w, unlock := WConn()
	defer unlock()

	err := w.conn.Exec(`
		PRAGMA analysis_limit=0;
		PRAGMA optimize;
	`)
	if err != nil {
		return fmt.Errorf("optimize db: %w", err)
	}

	return nil
}
