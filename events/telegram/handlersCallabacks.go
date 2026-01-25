package telegram

import (
	"gogobot/events/telegram/callbacks"
	"gogobot/events/telegram/types"
	"gogobot/events/telegram/types/botCommands"
	"gogobot/events/telegram/types/stateStorage"
	"strconv"
	"time"
)

const searchTTL = 2 * time.Minute

func (p *Processor) handlePageCallbacks(s *types.UserSession, cb callbacks.Callback) error {
	switch cb.Action {
	case "pick":
		return p.sendRandom(s, botCommands.PickMode)
	case "peek":
		return p.sendRandom(s, botCommands.PeekMode)
	default:
		return p.tgClient.SendMessage(s.ChatId, msgUnknownCommand)

	}
}

func (p *Processor) handleSearchCallbacks(s *types.UserSession, cb callbacks.Callback) error {
	switch cb.Action {
	case "pick":
		if isSearchExpired(s) {
			stateStorage.ResetSession(s)
			return p.tgClient.SendMessage(s.ChatId, "Search expired, please run /search again")
		}

		if s.LastSearchPages == nil {
			return p.tgClient.SendMessage(s.ChatId, "Search expired")
		}
		idx, err := strconv.Atoi(cb.Payload)
		if err != nil {
			return p.tgClient.SendMessage(s.ChatId, "Invalid selection")
		}
		if idx < 0 || idx >= len(s.LastSearchPages) {
			return p.tgClient.SendMessage(s.ChatId, "Index out of range")
		}

		page := s.LastSearchPages[idx]
		return p.tgClient.SendMessage(s.ChatId, page.URL)

	case "cancel":
		stateStorage.ResetSession(s)
		return p.tgClient.SendMessage(s.ChatId, "Search cancelled")
	default:
		return p.tgClient.SendMessage(s.ChatId, msgUnknownCommand)
	}
}

func (p *Processor) handleMenuCallbacks(s *types.UserSession, cb callbacks.Callback) error {
	switch cb.Action {
	case "help":
		return p.sendHelp(s)
	case "count":
		return p.sendCount(s)
	default:
		return p.tgClient.SendMessage(s.ChatId, msgUnknownCommand)

	}

}
