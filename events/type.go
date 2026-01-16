package events

type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}

type Processor interface {
	Process(event Event) error
}

type Type int

const (
	Unknown Type = iota
	Message
	Callback
)

type Event struct {
	Type    Type
	Meta    interface{}
	Payload interface{}
}
