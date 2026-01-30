package telegram

import (
	"gogobot/events/telegram/callbacks"
	"gogobot/events/telegram/types"
	"gogobot/events/telegram/types/botCommands"
	"gogobot/events/telegram/types/stateStorage"
	"strconv"
)

func (p *Processor) handlePageCallbacks(s *types.UserSession, cb callbacks.Callback) error {
	switch cb.Action {
	case "pick":
		return p.sendRandom(s, botCommands.PickMode)
	case "peek":
		return p.sendRandom(s, botCommands.PeekMode)
	case "cancel":
		stateStorage.ResetSession(s)
		return p.tgClient.SendMessage(s.ChatId, "Adding cancelled")
	default:
		return p.tgClient.SendMessage(s.ChatId, msgUnknownCommand)

	}
}

func (p *Processor) handleSearchCallbacks(s *types.UserSession, cb callbacks.Callback) error {
	switch cb.Action {
	case "pick":
		if s.LastSearchPages == nil {
			return NewUserError("Search expired")
		}
		idx, err := strconv.Atoi(cb.Payload)
		if err != nil {
			return NewUserError("Invalid selection")

		}
		if idx < 0 || idx >= len(s.LastSearchPages) {
			return NewUserError("Index out of range")
		}

		page := s.LastSearchPages[idx]
		return p.tgClient.SendMessage(s.ChatId, page.URL)

	case "cancel":
		stateStorage.ResetSession(s)
		return NewUserError("Search cancelled")
	default:
		return NewUserError(msgUnknownCommand)
	}
}

func (p *Processor) handleMenuCallbacks(s *types.UserSession, cb callbacks.Callback) error {
	switch cb.Action {
	case "help":
		return p.sendHelp(s)
	case "count":
		return p.sendCount(s)
	default:
		return NewUserError(msgUnknownCommand)

	}

}
