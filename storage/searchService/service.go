package searchService

import (
	"gogobot/lib/e"
	"gogobot/storage"
)

type service struct {
	storage storage.Storage
}

func New(storage storage.Storage) Service {
	return &service{storage: storage}
}

func (s *service) Search(userName string, tags []string) ([]*storage.Page, error) {
	if len(tags) == 0 {
		return nil, storage.ErrNoTags
	}

	pages, err := s.storage.FilterByTags(userName, tags)

	if err != nil {
		return nil, e.Wrap("search failed", err)
	}
	return pages, nil
}
