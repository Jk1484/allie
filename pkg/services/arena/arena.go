package arena

import (
	"allie/pkg/services/mage"

	"go.uber.org/fx"
)

var Module = fx.Provide(New)

type Service interface {
	GetMagesExcept(username string) []mage.Mage
	GetAllMages() []mage.Mage
	RemoveMage(username string)
	AddMage(m mage.Mage)
	GetByUsername(username string) *mage.Mage
}

type service struct {
	Mages map[string]*mage.Mage
}

type Params struct {
	fx.In
}

func New(p Params) Service {
	return &service{
		Mages: make(map[string]*mage.Mage),
	}
}

func (s *service) GetAllMages() []mage.Mage {
	return s.GetMagesExcept("")
}

func (s *service) GetMagesExcept(username string) []mage.Mage {
	mages := make([]mage.Mage, 0, len(s.Mages))
	for _, v := range s.Mages {
		if v.Username == username {
			continue
		}

		mages = append(mages, *v)
	}

	return mages
}

func (s *service) RemoveMage(username string) {
	delete(s.Mages, username)
}

func (s *service) AddMage(m mage.Mage) {
	s.Mages[m.Username] = &m
}

func (s *service) GetByUsername(username string) *mage.Mage {
	return s.Mages[username]
}
