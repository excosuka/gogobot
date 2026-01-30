package session

import "gogobot/events/telegram/types"

type ExpireHandler func(session *types.UserSession)

type Manager interface {
	Touch(session *types.UserSession)
	Stop(chatID int)
}
