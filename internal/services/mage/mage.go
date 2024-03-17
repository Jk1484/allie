package mage

import (
	"allie/internal/repositories/mage"
	"allie/internal/services/utils"
	"database/sql"
	"errors"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/lib/pq"
	"go.uber.org/fx"
)

var Module = fx.Provide(New)

type Service interface {
	Create(m Mage) error
	GetByUsername(string) (*Mage, error)
	UpdateHPByUsername(username string, newHP int) error
}

type service struct {
	mageRepository mage.Repository
}

type Params struct {
	fx.In
	MageRepository mage.Repository
}

func New(p Params) Service {
	return &service{
		mageRepository: p.MageRepository,
	}
}

type Mage struct {
	ID       int
	Username string
	Password string
	HP       int

	Conn *websocket.Conn
}

func (s *service) Create(m Mage) error {
	m.Username = strings.Trim(m.Username, " ")
	m.Password = strings.Trim(m.Password, " ")

	if m.Username == "" {
		return errors.New("empty username")
	}

	if m.Password == "" {
		return errors.New("empty password")
	}

	mR := mage.Mage{
		ID:       m.ID,
		Username: m.Username,
		Password: m.Password,
		HP:       m.HP,
	}

	err := s.mageRepository.Create(mR)
	if err != nil {
		v, ok := err.(*pq.Error)
		if !ok {
			return err
		}

		if v.Code == "23505" {
			return utils.ErrAlreadyExists
		}

		return err
	}

	return nil
}

func (s *service) GetByUsername(username string) (*Mage, error) {
	mr, err := s.mageRepository.GetByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}

		return nil, err
	}

	m := Mage{
		ID:       mr.ID,
		Username: mr.Username,
		Password: mr.Password,
		HP:       mr.HP,
	}

	return &m, nil
}

func (s *service) UpdateHPByUsername(username string, newHP int) error {
	return s.mageRepository.UpdateHPByUsername(username, newHP)
}
