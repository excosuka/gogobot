package telegram

import (
	"errors"
	"fmt"
	keyboards "gogobot/events/telegram/keyboards"
	"gogobot/lib/e"
	"gogobot/storage"
	"log"
	"net/url"
	"strconv"
	"strings"
)

const (
	HelpCmd   = "/help"
	StartCmd  = "/start"
	CountCmd  = "/count"
	PickMode  = "/pick"
	PeekMode  = "/peek"
	MenuMode  = "/menu"
	ListCmd   = "/list"
	SearchCmd = "/search"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)
	parts := strings.Fields(text)
	cmd := parts[0]

	tags := normalizeTags(parts)

	log.Printf("got new command: %s from %s with tags: %s", text, username, tags)

	if isAddCmd(cmd) {
		return p.savePage(chatID, text, username, tags)
	}

	switch cmd {
	case PickMode:
		return p.sendRandom(chatID, username, PickMode)
	case PeekMode:
		return p.sendRandom(chatID, username, PeekMode)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	case CountCmd:
		return p.sendCount(chatID, username)
	case ListCmd:
		return p.sendList(chatID, username)
	case SearchCmd:
		return p.sendFilteredByTag(chatID, username, tags)
	case MenuMode:
		return p.sendMenu(chatID, username)
	default:
		return p.tgClient.SendMessage(chatID, msgUnknownCommand)
	}
}

func (p *Processor) savePage(chatID int, pageURL string, username string, tags []string) (err error) {
	defer func() { err = e.WrapIfErr("can`t do command savePage()", err) }()

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
		Tags:     tags,
	}

	isExists, err := p.storage.IsExists(page)

	if err != nil {
		return err
	}

	if isExists {
		return p.tgClient.SendMessage(chatID, msgAlreadyExists)
	}
	if err := p.storage.Save(page); err != nil {
		return err
	}

	if err := p.tgClient.SendMessage(chatID, msgSaved); err != nil {
		return err
	}

	return nil

}

func (p *Processor) sendRandom(chatID int, username string, mode string) (err error) {
	defer func() { err = e.WrapIfErr("can`t do command sendRandom()", err) }()

	page, err := p.storage.PickRandom(username, mode)
	var messageToAnswer string

	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tgClient.SendMessage(chatID, msgNoSavedPages)
	}

	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	switch mode {
	case PickMode:
		messageToAnswer = page.URL + "\n(URL was deleted from storage)"
	case PeekMode:
		messageToAnswer = page.URL + "\n(URL wasn`t deleted from storage)"
	}

	if err := p.tgClient.SendMessage(chatID, messageToAnswer); err != nil {
		return err
	}

	return nil

}

func (p *Processor) sendCount(chatID int, userName string) (err error) {
	count, err := p.storage.Count(userName)

	if err != nil {
		return e.Wrap("can`t do command sendCount()", err)
	}

	messageToAnswer := msgCount + strconv.Itoa(count)

	if err := p.tgClient.SendMessage(chatID, messageToAnswer); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendList(chatID int, userName string) (err error) {
	listToMessage, err := p.storage.List(userName)

	if err != nil {
		return e.Wrap("can`t do command sendList()", err)
	}

	if len(listToMessage) == 0 {
		return p.tgClient.SendMessage(chatID, msgNoSavedPages)
	}

	var message string

	for i, file := range listToMessage {
		counter := strconv.Itoa(i + 1)
		if len(file.Tags) > 0 {
			message += fmt.Sprintf("Page %s: %s With Tags: %s \n", counter, file.URL, normalizeTags(file.Tags))
		} else {
			message += fmt.Sprintf("Page %s: %s \n", counter, file.URL)
		}
	}

	messageToAnswer := msgList + "\n " + message

	if err := p.tgClient.SendMessage(chatID, messageToAnswer); err != nil {
		return err
	}
	return nil
}

func (p *Processor) sendFilteredByTag(chatID int, userName string, tags []string) error {

	if len(tags) == 0 {
		return p.tgClient.SendMessage(chatID, msgEmptyHashTags)
	}

	filteredPages, err := p.storage.FilterByTags(userName, tags)
	if err != nil {
		return e.Wrap("can`t do command sendFilteredByTag()", err)
	}

	if len(filteredPages) == 0 {
		return p.tgClient.SendMessage(chatID, msgEmptyFilteredPages)
	}
	//fmt.Printf("[LOGS] getfiltered pages: %s \n", filteredPages)
	//fmt.Printf("[LOGS] tags to filter pages: %s \n", tags)

	var message string
	for i, page := range filteredPages {
		message += strconv.Itoa(i) + page.URL + "\n"
	}
	messageToAnswer := msgQuery + "\n " + message

	if err := p.tgClient.SendMessage(chatID, messageToAnswer); err != nil {
		return err
	}
	return nil

}

func (p *Processor) sendMenu(chatID int, username string) error {
	keyboard := keyboards.BuildMainMenuKeyboard()

	return p.tgClient.SendMessageWithKeyboard(chatID, "Main menu", keyboard)
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tgClient.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tgClient.SendMessage(chatID, msgHello)
}

func isAddCmd(text string) bool {
	return isURL(text)
}

// В будущем дописать
func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}

func normalizeTags(parts []string) []string {
	tags := make([]string, 0)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, "#") {
			tag := strings.TrimPrefix(part, "#")
			tag = strings.ToLower(tag)

			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}
