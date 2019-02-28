package errors

import "encoding/json"

type outputResource struct {
	Code string `json:"__type"`
	Msg  string `json:"message"`
}

func marshalOutput(c, m string) []byte {
	o := outputResource{Code: c, Msg: m}
	b, _ := json.Marshal(o)
	return b
}
