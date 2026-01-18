package keyboards

type ButtonCommand struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

var Pick = ButtonCommand{
	Text:         "Pick the article (with delete)",
	CallbackData: "/pick",
}
var Peek = ButtonCommand{
	Text:         "Peek the article (with no delete)",
	CallbackData: "/peek",
}

var Help = ButtonCommand{
	Text:         "Help",
	CallbackData: "/help",
}

var Count = ButtonCommand{
	Text:         "Count your articles in storage",
	CallbackData: "/count",
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
