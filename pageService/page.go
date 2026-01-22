package pageService

import "gogobot/storage"

type Service interface {
	Save(page *storage.Page) error
	List(userName string) ([]*storage.Page, error)
	Exists(page *storage.Page) (bool, error)
	Pick(userName string) (*storage.Page, error)
	Peek(userName string) (*storage.Page, error)
}
