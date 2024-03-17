package db

import (
	"allie/pkg/config"
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(New),
)

type Database interface {
	Close() error
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

type database struct {
	db      *sql.DB
	configs config.Configs
	limiter chan struct{} // limit amount of simultaneous calls to db
}

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Configs   config.Configs
}

func New(p Params) Database {
	newDB := &database{
		db:      connect(p.Configs),
		configs: p.Configs,
		limiter: make(chan struct{}, 10),
	}

	p.Lifecycle.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				return newDB.Close()
			},
		},
	)

	return newDB
}

func connect(cfg config.Configs) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Peek().Database.Host, cfg.Peek().Database.Port, cfg.Peek().Database.User, cfg.Peek().Database.Password, cfg.Peek().Database.Name)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	return db
}

func (d *database) Close() error {
	return d.db.Close()
}

func (d *database) QueryRow(query string, args ...interface{}) *sql.Row {
	d.Acquire()
	defer d.Release()
	return d.db.QueryRow(query, args...)
}

func (d *database) Exec(query string, args ...any) (sql.Result, error) {
	d.Acquire()
	defer d.Release()
	return d.db.Exec(query, args...)
}

func (d *database) Acquire() {
	d.limiter <- struct{}{}
}

func (d *database) Release() {
	<-d.limiter
}
