package telegram

import (
	"errors"
	"gogobot/events/telegram/callbacks"
	"gogobot/events/telegram/types"
)

type CallbackHandler func(*types.UserSession, callbacks.Callback) error

type CallbackRouter struct {
	handlers map[string]CallbackHandler
}

func NewCallbackRouter() *CallbackRouter {
	return &CallbackRouter{
		handlers: make(map[string]CallbackHandler),
	}
}

func (r *CallbackRouter) Register(key string, h CallbackHandler) {
	r.handlers[key] = h
}

func (r *CallbackRouter) Handle(
	p *Processor,
	s *types.UserSession,
	data string,
) error {

	cb := callbacks.Parse(data)

	key := cb.Domain + ":" + cb.Action

	h, ok := r.handlers[key]
	if !ok {
		return errors.New("unknown callback")
	}

	return h(s, cb)
}
