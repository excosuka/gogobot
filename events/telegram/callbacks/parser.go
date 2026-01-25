package callbacks

import "strings"

type Callback struct {
	Domain  string
	Action  string
	Payload string
}

func Parse(data string) Callback {
	parts := strings.Split(data, ":")

	cb := Callback{}

	if len(parts) > 0 {
		cb.Domain = parts[0]
	}
	if len(parts) > 1 {
		cb.Action = parts[1]
	}
	if len(parts) > 2 {
		cb.Payload = parts[2]
	}
	return cb
}
