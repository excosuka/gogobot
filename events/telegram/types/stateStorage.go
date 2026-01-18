package types

type StateStorage interface {
	Get(chatID int) (*UserSession, error)
	Save(session *UserSession) error
}

type UserState string

const (
	StateIdle           UserState = "idle"
	StateWaitingForUrl  UserState = "waiting_for_url"
	StateWaitingForTags UserState = "waiting_for_tags"
)

type UserSession struct {
	ChatId    int
	Username  string
	UserState UserState
	TempURL   string
}
