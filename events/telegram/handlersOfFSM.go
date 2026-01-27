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
)

func (p *Processor) handleIdle(s *types.UserSession, text string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command: %s from %s ", text, s.Username)

	if isAddCmd(text) {
		page := &storage.Page{
			URL:      text,
			UserName: s.Username,
		}

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
		return p.sendRandom(s, botCommands.PickMode)
	case botCommands.PeekMode:
		return p.sendRandom(s, botCommands.PeekMode)
	case botCommands.HelpCmd:
		return p.sendHelp(s)
	case botCommands.StartCmd:
		return p.sendHello(s)
	case botCommands.CountCmd:
		return p.sendCount(s)
	case botCommands.ListCmd:
		return p.sendList(s)
	case botCommands.SearchCmd:
		s.UserState = types.StateWaitingForTagsForSearch
		keyboard := keyboards.BuildCancelSearchKeyboard()
		return p.tgClient.SendMessageWithKeyboard(s.ChatId, msgToSpecifyHashTags, keyboard)

	case botCommands.MenuMode:
		return p.sendMenu(s)
	default:
		return p.tgClient.SendMessage(s.ChatId, msgUnknownCommand)

	}
}

func (p *Processor) handleWaitingForTagsForSearch(s *types.UserSession, text string) error {

	text = strings.TrimSpace(text)

	if strings.HasPrefix(text, "/") {

		return p.handleIdle(s, text)
	}

	tags := normalizeTags(strings.Fields(text))

	if len(tags) == 0 {
		return p.tgClient.SendMessage(s.ChatId, msgEmptyHashTags)
	}

	filteredPages, err := p.searchService.Search(s.Username, tags)
	if err != nil {
		return e.Wrap("can`t do command sendFilteredByTag()", err)
	}

	if len(filteredPages) == 0 {
		stateStorage.ResetSession(s)
		return p.tgClient.SendMessage(s.ChatId, msgEmptyFilteredPages)

	}

	var message string
	for i, page := range filteredPages {
		message += strconv.Itoa(i+1) + ". " + page.URL + "\n"
	}
	messageToAnswer := msgQuery + "\n " + message

	s.LastSearchTags = tags
	s.LastSearchPages = filteredPages

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
		return p.tgClient.SendMessage(s.ChatId, msgEmptyHashTags)
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
