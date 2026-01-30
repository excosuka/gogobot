package telegram

import (
	"gogobot/events/telegram/keyboards"
	"gogobot/events/telegram/types"
	"gogobot/events/telegram/types/botCommands"
	"gogobot/events/telegram/types/stateStorage"
	"gogobot/lib/e"
	"gogobot/storage"
	"log"
	"strconv"
	"strings"
	"time"
)

func (p *Processor) handleIdle(s *types.UserSession, text string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command: %s from %s ", text, s.Username)

	if isAddCmd(text) {
		page := &storage.Page{
			URL:      text,
			UserName: s.Username,
		}

		addAction(s, "Added "+text)

		isExists, err := p.pageService.Exists(page)
		if err != nil {
			return err
		}

		if isExists {
			return p.tgClient.SendMessage(s.ChatId, msgAlreadyExists)
		}

		s.TempURL = text
		s.UserState = types.StateWaitingForTags

		cancelKeyboard := keyboards.BuildCancelAddKeyboard()
		return p.tgClient.SendMessageWithKeyboard(s.ChatId, "Waiting for hashtags", cancelKeyboard)
	}

	switch text {
	case botCommands.PickMode:
		addAction(s, "completed command "+botCommands.PickMode)
		return p.sendRandom(s, botCommands.PickMode)
	case botCommands.PeekMode:
		addAction(s, "completed command "+botCommands.PeekMode)
		return p.sendRandom(s, botCommands.PeekMode)
	case botCommands.HelpCmd:
		addAction(s, "completed command "+botCommands.HelpCmd)
		return p.sendHelp(s)
	case botCommands.StartCmd:
		addAction(s, "completed command "+botCommands.StartCmd)
		return p.sendHello(s)
	case botCommands.CountCmd:
		addAction(s, "completed command "+botCommands.CountCmd)
		return p.sendCount(s)
	case botCommands.ListCmd:
		addAction(s, "completed command "+botCommands.ListCmd)
		return p.sendList(s)
	case botCommands.SearchCmd:
		addAction(s, "completed command "+botCommands.SearchCmd)
		s.UserState = types.StateWaitingForTagsForSearch
		keyboard := keyboards.BuildCancelSearchKeyboard()
		return p.tgClient.SendMessageWithKeyboard(s.ChatId, msgToSpecifyHashTags, keyboard)

	case botCommands.MenuMode:
		return p.sendMenu(s)
	default:
		return NewUserError(msgUnknownCommand)

	}
}

func (p *Processor) handleWaitingForTagsForSearch(s *types.UserSession, text string) error {

	text = strings.TrimSpace(text)

	if strings.HasPrefix(text, "/") {

		return p.handleIdle(s, text)
	}

	tags := normalizeTags(strings.Fields(text))

	if len(tags) == 0 {
		return NewUserError(msgEmptyHashTags)
	}

	filteredPages, err := p.searchService.Search(s.Username, tags)
	if err != nil {
		return e.Wrap("can`t do command sendFilteredByTag()", err)
	}

	if len(filteredPages) == 0 {
		stateStorage.ResetSession(s)
		return NewUserError(msgEmptyFilteredPages)

	}

	var message string
	for i, page := range filteredPages {
		message += strconv.Itoa(i+1) + ". " + page.URL + "\n"
	}
	messageToAnswer := msgQuery + "\n " + message

	s.LastSearchTags = tags
	s.LastSearchPages = filteredPages
	s.LastSearchAt = time.Now()
	stateStorage.ResetSession(s)

	kb := keyboards.BuildSearchResultKeyboard(filteredPages)

	if err := p.tgClient.SendMessageWithKeyboard(
		s.ChatId,
		messageToAnswer,
		kb,
	); err != nil {
		return err
	}

	return nil

}

func (p *Processor) handleWaitingForTags(s *types.UserSession, text string) (err error) {
	defer func() { err = e.WrapIfErr("can`t set hashtags to page", err) }()

	if strings.HasPrefix(text, "/") {
		return p.tgClient.SendMessage(s.ChatId,
			"You maybe have a mistake, now you are in hashtags input mode.\nIf you wanna go out use: /cancel")
	}
	tags := normalizeTags(strings.Fields(text))

	if len(tags) == 0 {
		return NewUserError(msgEmptyHashTags)
	}

	page := &storage.Page{
		URL:      s.TempURL,
		Tags:     tags,
		UserName: s.Username,
	}

	if err = p.pageService.Save(page); err != nil {
		return err
	}

	stateStorage.ResetSession(s)

	if err = p.tgClient.SendMessage(s.ChatId, msgSaved); err != nil {
		return err
	}

	return nil

}

func isSearchExpired(s *types.UserSession) bool {
	if s.LastSearchAt.IsZero() {
		return true
	}

	return time.Since(s.LastSearchAt) > searchTTL
}
