package telegram

import (
	"errors"
	"gogobot/clients/telegram"
	"gogobot/events"
	"gogobot/events/telegram/callbacks"
	"gogobot/events/telegram/keyboards"
	"gogobot/events/telegram/types"
	"gogobot/events/telegram/types/botCommands"
	"gogobot/events/telegram/types/stateStorage"
	"gogobot/lib/e"
	"gogobot/storage/pageService"
	"gogobot/storage/searchService"
	"strconv"
	"time"
)

const sessionTTL = 60 * time.Minute

type Processor struct {
	tgClient        *telegram.Client
	offset          int
	pageService     pageService.Service
	stateStorage    *stateStorage.StateStorage
	searchService   searchService.Service
	callbacksRouter *CallbackRouter
}

type Meta struct {
	ChatID   int
	Username string
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
	ErrUnknownState     = errors.New("unknown state")
)

func New(client *telegram.Client, pageService pageService.Service, stateStorage *stateStorage.StateStorage, searchService searchService.Service, router *CallbackRouter) *Processor {
	p := &Processor{
		tgClient:        client,
		pageService:     pageService,
		stateStorage:    stateStorage,
		searchService:   searchService,
		callbacksRouter: router,
	}
	p.registerCallbacks()
	return p

}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := p.tgClient.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("cannot fetch events", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, u := range updates {
		res = append(res, event(u))
	}

	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

func (p *Processor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		payload := event.Payload.(types.MessagePayload)
		meta := event.Meta.(Meta)

		session, err := p.stateStorage.Get(meta.ChatID)
		if session == nil {
			session = &types.UserSession{
				ChatId:       meta.ChatID,
				Username:     meta.Username,
				UserState:    types.StateIdle,
				LastActivity: time.Now(),
			}
		}

		if p.checkSessionTimeout(session) {
			return nil
		}

		session.LastActivity = time.Now()

		err = p.handleMessage(session, payload.Text)

		if err != nil {
			return e.Wrap("cannot process message", err)
		}

		return p.stateStorage.Save(session)
	case events.Callback:
		payload := event.Payload.(types.CallbackPayload)
		meta := event.Meta.(Meta)

		session, err := p.stateStorage.Get(meta.ChatID)

		if session == nil {
			session = &types.UserSession{
				ChatId:       meta.ChatID,
				Username:     meta.Username,
				UserState:    types.StateIdle,
				LastActivity: time.Now(),
			}

		}

		if p.checkSessionTimeout(session) {
			return nil
		}

		session.LastActivity = time.Now()

		if err != nil {
			return e.Wrap("cannot process callback", err)
		}

		return p.processCallbacks(session, payload.Data)
	default:
		return e.Wrap("cannot process message", ErrUnknownEventType)

	}
}

func (p *Processor) processCallbacks(
	s *types.UserSession,
	data string,
) error {
	return p.callbacksRouter.Handle(p, s, data)
}

func event(upd telegram.Update) events.Event {
	if upd.Message != nil {
		return events.Event{
			Type: events.Message,
			Meta: Meta{
				ChatID:   upd.Message.Chat.ID,
				Username: upd.Message.From.Username,
			},
			Payload: types.MessagePayload{
				Text: upd.Message.Text,
			},
		}
	}

	if upd.CallbackQuery != nil {
		cq := upd.CallbackQuery
		chatID := 0
		username := cq.From.Username

		if cq.Message != nil {
			chatID = cq.Message.Chat.ID
		}

		return events.Event{
			Type: events.Callback,
			Meta: Meta{
				ChatID:   chatID,
				Username: username,
			},
			Payload: types.CallbackPayload{
				Data: cq.Data,
			},
		}
	}

	return events.Event{
		Type: events.Unknown,
	}
}

func (p *Processor) checkSessionTimeout(s *types.UserSession) bool {
	if time.Since(s.LastActivity) <= sessionTTL {
		return false
	}

	wasInDialog := s.UserState != types.StateIdle

	stateStorage.ResetSession(s)

	if wasInDialog {
		_ = p.tgClient.SendMessage(
			s.ChatId,
			"⌛ Диалог был сброшен из-за неактивности",
		)
	}

	return true
}

func (p *Processor) registerCallbacks() {
	p.callbacksRouter.Register("page:pick", p.cbPick)
	p.callbacksRouter.Register("page:peek", p.cbPeek)
	p.callbacksRouter.Register("menu:count", p.cbCount)

	p.callbacksRouter.Register("search:pick", p.cbSearchPick)
	p.callbacksRouter.Register("search:repeat", p.cbSearchRepeat)
}

func (p *Processor) cbPick(
	s *types.UserSession,
	_ callbacks.Callback,
) error {
	return p.sendRandom(s, botCommands.PickMode)
}

func (p *Processor) cbPeek(session *types.UserSession, _ callbacks.Callback) error {
	return p.sendRandom(session, botCommands.PeekMode)
}

func (p *Processor) cbCount(session *types.UserSession, _ callbacks.Callback) error {
	return p.sendCount(session)
}

func (p *Processor) cbSearchRepeat(
	s *types.UserSession,
	_ callbacks.Callback,
) error {

	// 1. Проверяем, есть ли вообще результаты
	if len(s.LastSearchPages) == 0 {
		return p.tgClient.SendMessage(
			s.ChatId,
			"Нет активного поиска. Используй /search",
		)
	}

	// 2. Проверяем TTL
	if isSearchExpired(s) {
		stateStorage.ResetSession(s)
		return p.tgClient.SendMessage(
			s.ChatId,
			"Поиск устарел. Запусти /search заново",
		)
	}

	// 3. Собираем сообщение
	msg := listPagesToMessage(s.LastSearchPages)

	// 4. Клавиатура
	kb := keyboards.BuildSearchResultKeyboard(s.LastSearchPages)

	// 5. Отправляем
	return p.tgClient.SendMessageWithKeyboard(
		s.ChatId,
		msg,
		kb,
	)
}

func (p *Processor) cbSearchPick(
	s *types.UserSession,
	cb callbacks.Callback,
) error {

	if isSearchExpired(s) || len(s.LastSearchPages) == 0 {
		stateStorage.ResetSession(s)
		s.LastActivity = time.Now()
		return p.tgClient.SendMessage(
			s.ChatId,
			"Search expired, run /search again",
		)
	}

	idx, err := strconv.Atoi(cb.Payload)
	if err != nil || idx < 0 || idx >= len(s.LastSearchPages) {
		return p.tgClient.SendMessage(s.ChatId, "Invalid selection")
	}

	page := s.LastSearchPages[idx]
	return p.tgClient.SendMessage(s.ChatId, page.URL)
}
