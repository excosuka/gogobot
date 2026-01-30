package telegram

import (
	"errors"
	"gogobot/clients/telegram"
	"gogobot/events"
	"gogobot/events/session"
	"gogobot/events/telegram/callbacks"
	"gogobot/events/telegram/types"
	"gogobot/events/telegram/types/stateStorage"
	"gogobot/lib/e"
	"gogobot/storage/pageService"
	"gogobot/storage/searchService"
	"time"
)

const searchTTL = 10 * time.Minute

type Processor struct {
	tgClient      *telegram.Client
	offset        int
	pageService   pageService.Service
	stateStorage  *stateStorage.StateStorage
	searchService searchService.Service
	sessionManger session.Manager
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

func New(
	client *telegram.Client,
	pageService pageService.Service,
	stateStorage *stateStorage.StateStorage,
	searchService searchService.Service,
	sessionMgr session.Manager,
) *Processor {
	return &Processor{
		tgClient:      client,
		pageService:   pageService,
		stateStorage:  stateStorage,
		searchService: searchService,
		sessionManger: sessionMgr,
	}
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
				ChatId:    meta.ChatID,
				Username:  meta.Username,
				UserState: types.StateIdle,
			}
		}

		p.sessionManger.Touch(session)

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
				ChatId:    meta.ChatID,
				Username:  meta.Username,
				UserState: types.StateIdle,
			}

		}

		p.sessionManger.Touch(session)

		if err != nil {
			return e.Wrap("cannot process callback", err)
		}

		err = p.processCallbacks(session, payload.Data)
		if err != nil {
			return err
		}

		return p.stateStorage.Save(session)
	default:
		return e.Wrap("cannot process message", ErrUnknownEventType)

	}
}

func (p *Processor) processCallbacks(s *types.UserSession, data string) error {
	cb := callbacks.Parse(data)
	switch cb.Domain {
	case "page":
		return p.handlePageCallbacks(s, cb)
	case "search":
		return p.handleSearchCallbacks(s, cb)
	case "menu":
		return p.handleMenuCallbacks(s, cb)
	default:
		return NewUserError(msgUnknownCommand)

	}

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
