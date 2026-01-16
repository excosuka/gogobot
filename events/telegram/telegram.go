package telegram

import (
	"errors"
	"gogobot/clients/telegram"
	"gogobot/events"
	"gogobot/events/telegram/types"
	"gogobot/lib/e"
	"gogobot/storage"
)

type Processor struct {
	tgClient *telegram.Client
	offset   int
	storage  storage.Storage
}

type Meta struct {
	ChatID   int
	Username string
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

func New(client *telegram.Client, storage storage.Storage) *Processor {
	return &Processor{
		tgClient: client,
		storage:  storage,
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
		return p.doCmd(payload.Text, meta.ChatID, meta.Username)
	case events.Callback:
		payload := event.Payload.(types.CallbackPayload)
		meta := event.Meta.(Meta)
		return p.processCallbacks(meta.ChatID, payload.Data, meta.Username)
	default:
		return e.Wrap("cannot process message", ErrUnknownEventType)

	}
}

//func (p *Processor) processMessage(event events.Event) error {
//	metaFromEvent, err := meta(event)
//	if err != nil {
//		return e.Wrap("cant process message", err)
//	}
//
//	if err := p.doCmd(event.Text, metaFromEvent.ChatID, metaFromEvent.Username); err != nil {
//		return e.Wrap("cant process message", err)
//	}
//	return nil
//}

func (p *Processor) processCallbacks(chatID int, callbackData string, username string) error {
	switch callbackData {
	case "/pick":
		return p.sendRandom(chatID, username, PickMode)
	case "/peek":
		return p.sendRandom(chatID, username, PeekMode)
	case "/count":
		return p.sendCount(chatID, username)
	case "/help":
		return p.sendHelp(chatID)
	}
	return p.tgClient.SendMessage(chatID, msgUnknownCommand)

}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("cannot extract meta from message", ErrUnknownMetaType)
	}
	return res, nil
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

func fetchType(upd telegram.Update) events.Type {
	if upd.Message == nil {
		return events.Unknown
	}
	return events.Message
}

func fetchText(upd telegram.Update) string {
	if upd.Message == nil {
		return ""

	}
	return upd.Message.Text
}
