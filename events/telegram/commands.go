package telegram

import (
	"errors"
	keyboards "gogobot/events/telegram/keyboards"
	"gogobot/lib/e"
	"gogobot/storage"
	"log"
	"net/url"
	"strconv"
	"strings"
)

const (
	HelpCmd = "/help"

	StartCmd = "/start"
	CountCmd = "/count"

	PickMode = "/pick"
	PeekMode = "/peek"
	MenuMode = "/menu"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command: %s from %s", text, username)

	//add page: http://...
	// rnd page: /rnd
	// help: /help
	// start: /start: hi + help

	if isAddCmd(text) {
		return p.savePage(chatID, text, username)
	}

	switch text {
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
	case MenuMode:
		return p.sendMenu(chatID, username)
	default:
		return p.tgClient.SendMessage(chatID, msgUnknownCommand)
	}
}

func (p *Processor) savePage(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("can`t do command savePage()", err) }()

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
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

	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tgClient.SendMessage(chatID, msgNoSavedPages)
	}

	if err := p.tgClient.SendMessage(chatID, page.URL); err != nil {
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
