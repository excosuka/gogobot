package pageService

import (
	"gogobot/events/telegram"
	"gogobot/storage"
)

type service struct {
	storage storage.Storage
}

func New(storage storage.Storage) Service {
	return &service{storage: storage}
}

func (s *service) Save(p *storage.Page) error {
	exists, err := s.storage.IsExists(p)
	if err != nil {
		return err
	}

	if exists {
		return storage.ErrPageExists
	}

	return s.storage.Save(p)
}

func (s *service) List(userName string) ([]*storage.Page, error) {
	return s.storage.List(userName)
}

func (s *service) Exists(p *storage.Page) (bool, error) {
	return s.storage.IsExists(p)
}

func (s *service) Pick(userName string) (*storage.Page, error) {
	return s.storage.PickRandom(userName, telegram.PickMode)
}

func (s *service) Peek(userName string) (*storage.Page, error) {
	return s.storage.PickRandom(userName, telegram.PeekMode)
}
