package telegram

import (
	"context"
	"errors"
	"gogobot/events"
	keyboards "gogobot/events/telegram/keyboards"
	"gogobot/events/telegram/types"
	"gogobot/events/telegram/types/botCommands"
	"gogobot/events/telegram/types/stateStorage"
	"gogobot/lib/e"
	"gogobot/storage"
	"math/rand"
	"strconv"
	"time"
)

func (p *Processor) handleMessage(s *types.UserSession, text string) error {
	if text == botCommands.CancelCmd {
		stateStorage.ResetSession(s)
		return p.tgClient.SendMessage(s.ChatId, msgCanceled)
	}

	switch s.UserState {

	case types.StateIdle:
		return p.handleIdle(s, text)

	case types.StateWaitingForTags:
		return p.handleWaitingForTags(s, text)

	case types.StateWaitingForTagsForSearch:
		if isSearchExpired(s) || len(s.LastSearchPages) == 0 {
			stateStorage.ResetSession(s)
			return p.tgClient.SendMessage(s.ChatId, msgZeroSearched)
		}
		return p.handleWaitingForTagsForSearch(s, text)

	default:
		return ErrUnknownState
	}
}

func (p *Processor) sendRandom(s *types.UserSession, mode string) (err error) {
	defer func() { err = e.WrapIfErr("can`t do command sendRandom()", err) }()

	var messageToAnswer string

	switch mode {
	case botCommands.PickMode:
		page, err := p.pageService.Pick(s.Username)
		if errors.Is(err, storage.ErrNoSavedPages) {
			return p.tgClient.SendMessage(s.ChatId, msgNoSavedPages)
		}

		if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
			return err
		}

		messageToAnswer = "🎯 Here your url:\n" + page.URL + "\n(URL was deleted from storage)"

	case botCommands.PeekMode:

		page, err := p.pageService.Peek(s.Username)
		if errors.Is(err, storage.ErrNoSavedPages) {
			return p.tgClient.SendMessage(s.ChatId, msgNoSavedPages)
		}

		if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
			return err
		}

		messageToAnswer = "👀 Here your url:\n" + page.URL + "\n(URL wasn`t deleted from storage)"

	}
	if err = p.tgClient.SendMessage(s.ChatId, messageToAnswer); err != nil {
		return err
	}

	return nil

}

func (p *Processor) sendCount(s *types.UserSession) (err error) {
	count, err := p.pageService.Count(s.Username)

	if err != nil {
		return e.Wrap("can`t do command sendCount()", err)
	}

	messageToAnswer := msgCount + strconv.Itoa(count)

	if err := p.tgClient.SendMessage(s.ChatId, messageToAnswer); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendList(s *types.UserSession) (err error) {
	listToMessage, err := p.pageService.List(s.Username)

	if err != nil {
		return e.Wrap("can`t do command sendList()", err)
	}
	if len(listToMessage) == 0 {
		return p.tgClient.SendMessage(s.ChatId, msgNoSavedPages)
	}
	messageToAnswer := listPagesToMessage(listToMessage)

	if err := p.tgClient.SendMessage(s.ChatId, messageToAnswer); err != nil {
		return err
	}
	return nil
}
func (p *Processor) sendMenu(s *types.UserSession) error {
	keyboard := keyboards.BuildMainMenuKeyboard()
	_, err := p.tgClient.SendMessageWithKeyboard(s.ChatId, "Main menu", keyboard)
	return err
}

func (p *Processor) sendHelp(s *types.UserSession) error {
	keyboardForHelp := keyboards.BuildMainMenuKeyboard()
	_, err := p.tgClient.SendMessageWithKeyboard(s.ChatId, msgHelp, keyboardForHelp)
	return err
}

func (p *Processor) sendHello(s *types.UserSession) error {
	return p.tgClient.SendMessage(s.ChatId, msgHello)
}

func (p *Processor) repeatSearch(s *types.UserSession) error {
	if len(s.LastSearchPages) == 0 {
		return NewUserError("No previous search")
	}
	pages, err := p.searchService.Search(s.Username, s.LastSearchTags)
	if err != nil {
		return err
	}
	s.LastSearchPages = pages
	_, err = p.tgClient.SendMessageWithKeyboard(
		s.ChatId,
		listPagesToMessage(pages),
		keyboards.BuildSearchResultKeyboard(pages),
	)
	return err
}

func (p *Processor) pickFromSearch(s *types.UserSession) error {
	if len(s.LastSearchPages) == 0 {
		return NewUserError("No search results")
	}

	page := s.LastSearchPages[rand.Intn(len(s.LastSearchPages))]

	return p.tgClient.SendMessage(
		s.ChatId,
		page.URL,
	)
}

func (p *Processor) SendUserError(
	event events.Event,
	err UserError,
) error {

	meta := event.Meta.(Meta)

	return p.tgClient.SendMessage(
		meta.ChatID,
		"⚠️ "+err.Message,
	)
}

func (p *Processor) parseAndPreview(chatID int, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	page, err := p.parserService.Parse(ctx, url)
	if err != nil {
		_ = p.tgClient.SendMessage(chatID, "❌ Не удалось разобрать страницу")
		return
	}

	s, err := p.stateStorage.Get(chatID)
	if err != nil || s == nil {
		return
	}

	if s.UserState != types.StateIdle {
		_ = p.tgClient.SendMessage(chatID, "ℹ️ Парсинг отменён")
		return
	}

	s.TempURL = url
	s.ParsedPage = page
	s.UserState = types.StateWaitingForTags

	if err := p.stateStorage.Save(s); err != nil {
		return
	}

	_, _ = p.tgClient.SendMessageWithKeyboard(
		chatID,
		renderPreview(page),
		keyboards.BuildParseConfirmKeyboard(),
	)
}
