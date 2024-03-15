package db

import (
	"allie/configs"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(New),
	fx.Provide(connect),
)

type Database interface {
	Connection() *sql.DB
	CloseConnection() error
}

type database struct {
	db      *sql.DB
	configs configs.Configs
}

type Params struct {
	fx.In
	Configs configs.Configs
	DB      *sql.DB
}

func New(p Params) Database {
	return &database{
		db:      p.DB,
		configs: p.Configs,
	}
}

func connect(cfg configs.Configs) *sql.DB {
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

func (d *database) Connection() *sql.DB {
	return d.db
}

func (d *database) CloseConnection() error {
	return d.db.Close()
}
