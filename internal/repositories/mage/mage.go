package mage

import (
	"allie/internal/db"

	"go.uber.org/fx"
)

var Module = fx.Provide(New)

type Repository interface {
	Create(m Mage) error
	GetByUsername(username string) (*Mage, error)
	ReduceHPByUsername(username string) error
}

type repository struct {
	db db.Database
}

type Params struct {
	fx.In
	DB db.Database
}

func New(p Params) Repository {
	return &repository{
		db: p.DB,
	}
}

type Mage struct {
	ID       int
	Username string
	Password string
	HP       int
}

func (r *repository) Create(m Mage) error {
	query := `
		INSERT INTO mages(username, password)
		VALUES($1, $2)
	`

	_, err := r.db.Connection().Exec(query, m.Username, m.Password)

	return err
}

func (r *repository) GetByUsername(username string) (*Mage, error) {
	var m Mage

	query := `
		SELECT id, username, password, hp
		FROM mages
		WHERE username = $1
	`

	err := r.db.Connection().QueryRow(query, username).Scan(&m.ID, &m.Username, &m.Password, &m.HP)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *repository) ReduceHPByUsername(username string) error {
	query := `
		UPDATE mages
		SET hp = hp-10
		WHERE username = $1
	`

	// todo: check if affected
	_, err := r.db.Connection().Exec(query, username)

	return err
}
