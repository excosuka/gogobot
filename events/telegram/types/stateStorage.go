package types

import (
	"gogobot/storage"
	"time"
)

type StateStorage interface {
	Get(chatID int) (*UserSession, error)
	Save(session *UserSession) error
}

type UserState string

const (
	StateIdle                    UserState = "idle"
	StateWaitingForTags          UserState = "waiting_for_tags"
	StateWaitingForTagsForSearch UserState = "waiting_for_tags_for_search"
)

type UserSession struct {
	ChatId       int
	Username     string
	UserState    UserState
	TempURL      string
	LastActivity time.Time

	LastActions     []string
	LastSearchTags  []string
	LastSearchPages []*storage.Page
	LastSearchAt    time.Time
	SearchStartedAt time.Time
}
