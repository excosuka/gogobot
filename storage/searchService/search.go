package searchService

import "gogobot/storage"

type Service interface {
	Search(userName string, tags []string) ([]*storage.Page, error)
}
