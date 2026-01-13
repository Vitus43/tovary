package events

import "encoding/json"

type ErrorEvent struct {
	Location string
	Msg      string
}

func MustMarshal(location, msg string) []byte {
	data, err := json.Marshal(ErrorEvent{Location: location, Msg: msg})
	if err != nil {
		panic(err)
	}

	return data
}
