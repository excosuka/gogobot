package session

import (
	"gogobot/events/telegram/types"
	"sync"
	"time"
)

type TtlManager struct {
	ttl      time.Duration
	onExpire ExpireHandler

	mu     sync.Mutex
	timers map[int]*time.Timer
}

func NewTTLManager(ttl time.Duration, onExpire ExpireHandler) Manager {
	return &TtlManager{
		ttl:      ttl,
		onExpire: onExpire,
		timers:   make(map[int]*time.Timer),
	}
}

func (m *TtlManager) Touch(s *types.UserSession) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldTimer, ok := m.timers[s.ChatId]; ok {
		oldTimer.Stop()
	}
	m.timers[s.ChatId] = time.AfterFunc(m.ttl, func() {
		m.mu.Lock()
		delete(m.timers, s.ChatId)
		m.mu.Unlock()
		m.onExpire(s)
	})

}

func (m *TtlManager) Stop(chatId int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if t, ok := m.timers[chatId]; ok {
		t.Stop()
		delete(m.timers, chatId)
	}
}
