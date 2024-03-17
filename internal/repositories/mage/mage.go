package mage

import (
	"allie/internal/db"

	"go.uber.org/fx"
)

var Module = fx.Provide(New)

type Repository interface {
	Create(m Mage) error
	GetByUsername(username string) (*Mage, error)
	UpdateHPByUsername(username string, newHP int) error
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

	_, err := r.db.Exec(query, m.Username, m.Password)

	return err
}

func (r *repository) GetByUsername(username string) (*Mage, error) {
	var m Mage

	query := `
		SELECT id, username, password, hp
		FROM mages
		WHERE username = $1
	`

	err := r.db.QueryRow(query, username).Scan(&m.ID, &m.Username, &m.Password, &m.HP)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *repository) UpdateHPByUsername(username string, newHP int) error {
	query := `
		UPDATE mages
		SET hp = $1
		WHERE username = $2
	`

	// todo: check if affected
	_, err := r.db.Exec(query, newHP, username)

	return err
}
