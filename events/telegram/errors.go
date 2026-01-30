package telegram

import "gogobot/events"

type UserError struct {
	Message string
}

func (e UserError) Error() string {
	return e.Message
}

func NewUserError(msg string) UserError {
	return UserError{
		Message: msg,
	}
}

func (e UserError) Send(
	p *Processor,
	event events.Event,
) error {
	meta := event.Meta.(Meta)
	return p.tgClient.SendMessage(
		meta.ChatID,
		"⚠️ "+e.Message,
	)
}
