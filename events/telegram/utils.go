package telegram

import (
	"fmt"
	"gogobot/events/telegram/types"
	"gogobot/parserService"
	"gogobot/storage"
	"net/url"
	"strconv"
	"strings"
)

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

func listPagesToMessage(parts []*storage.Page) string {

	var message string

	for i, file := range parts {
		counter := strconv.Itoa(i + 1)
		if len(file.Tags) > 0 {
			message += fmt.Sprintf("Page %s: %s With Tags: %s \n", counter, file.URL, file.Tags)
		} else {
			message += fmt.Sprintf("Page %s: %s \n", counter, file.URL)
		}
	}

	messageToAnswer := msgList + "\n " + message

	return messageToAnswer

}

func addAction(s *types.UserSession, action string) {
	s.LastActions = append(s.LastActions, action)

	if len(s.LastActions) > 5 {
		s.LastActions = s.LastActions[1:]
	}
}

func renderPreview(p *parserService.ParsedPage) string {
	msg := ""
	if p.Title != "" {
		msg += "📎 " + p.Title + "\n"
	}
	if p.Description != "" {
		msg += p.Description + "\n"
	}
	if p.SiteName != "" {
		msg += "🌐 " + p.SiteName + "\n"
	}
	msg += p.URL
	return msg
}
