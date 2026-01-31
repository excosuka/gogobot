package keyboards

import (
	"gogobot/storage"
	"strconv"
)

type ButtonCommand struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

var Pick = ButtonCommand{
	Text:         "🎲 Pick article (delete)",
	CallbackData: "page:pick",
}
var Peek = ButtonCommand{
	Text:         "🎲 Pick article (keep)",
	CallbackData: "page:peek",
}

var Help = ButtonCommand{
	Text:         "ℹ️ Help",
	CallbackData: "menu:help",
}

var Count = ButtonCommand{
	Text:         "📊 Articles count",
	CallbackData: "menu:count",
}

var CancelSearch = ButtonCommand{
	Text:         "❌ Cancel to Search",
	CallbackData: "search:cancel",
}

var CancelAdd = ButtonCommand{
	Text:         "❌ Cancel to add link",
	CallbackData: "page:cancel",
}

var SearchRepeat = ButtonCommand{
	Text:         "🔁 Repeat search",
	CallbackData: "search:repeat",
}

var SearchPick = ButtonCommand{
	Text:         "🎯 Pick from results",
	CallbackData: "search:pick",
}

type ReplyMenuKeyboard struct {
	InlineKeyboard [][]ButtonCommand `json:"inline_keyboard"`
}

func BuildMainMenuKeyboard() ReplyMenuKeyboard {
	return ReplyMenuKeyboard{
		InlineKeyboard: [][]ButtonCommand{
			{Pick, Peek},
			{Help, Count},
		},
	}
}

func BuildSearchResultKeyboard(pages []*storage.Page) ReplyMenuKeyboard {
	rows := make([][]ButtonCommand, 0)

	for i := range pages {
		rows = append(rows, []ButtonCommand{
			{
				Text:         strconv.Itoa(i + 1),
				CallbackData: "search:pick:" + strconv.Itoa(i),
			},
		})
	}

	rows = append(rows, []ButtonCommand{
		{
			Text:         "❌ Cancel",
			CallbackData: "search:cancel",
		},
	})

	return ReplyMenuKeyboard{
		InlineKeyboard: rows,
	}
}

func BuildCancelSearchKeyboard() ReplyMenuKeyboard {
	return ReplyMenuKeyboard{
		InlineKeyboard: [][]ButtonCommand{{CancelSearch}},
	}
}

func BuildCancelAddKeyboard() ReplyMenuKeyboard {
	return ReplyMenuKeyboard{
		InlineKeyboard: [][]ButtonCommand{{CancelAdd}},
	}
}

func BuildParseConfirmKeyboard() ReplyMenuKeyboard {
	return ReplyMenuKeyboard{
		InlineKeyboard: [][]ButtonCommand{
			{
				{
					Text:         "✅ Сохранить",
					CallbackData: "parse:save",
				},
				{
					Text:         "✏️ Изменить теги",
					CallbackData: "parse:edit",
				},
				{
					Text:         "❌ Отмена",
					CallbackData: "parse:cancel",
				},
			},
		},
	}
}
