package stateStorage

import (
	"errors"
	"gogobot/events/telegram/types"
)

var ErrSessionNotFound = errors.New("session not found")

type StateStorage map[int]*types.UserSession

func New() *StateStorage {
	storage := make(StateStorage)
	return &storage
}

func (s *StateStorage) Get(chatID int) (*types.UserSession, error) {
	session, ok := (*s)[chatID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *StateStorage) Save(session *types.UserSession) error {

	chatID := session.ChatId
	(*s)[chatID] = session
	return nil

}
