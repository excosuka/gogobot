package storage

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"gogobot/lib/e"
	"io"
)

type Storage interface {
	Save(p *Page) error
	PickRandom(userName string, mode string) (*Page, error)
	Remove(p *Page) error
	IsExists(p *Page) (bool, error)
	Count(userName string) (int, error)
	List(userName string) ([]*Page, error)
	FilterByTags(userName string, tagsToCheck []string) ([]*Page, error)
}

var ErrNoSavedPages = errors.New("no saved pages")
var ErrPageExists = errors.New("page exists")
var ErrNoHaveStorage = errors.New("no storage available")
var ErrNoTags = errors.New("no tags in request")

type Page struct {
	URL      string
	UserName string
	Tags     []string
}

func (p *Page) Hash() (string, error) {
	h := sha1.New()
	if _, err := io.WriteString(h, p.URL); err != nil {
		return "", e.Wrap("can`t calculate hash", err)
	}

	if _, err := io.WriteString(h, p.UserName); err != nil {
		return "", e.Wrap("can`t calculate hash", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
