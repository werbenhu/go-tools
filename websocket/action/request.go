package action

import "git.aimore.com/golang/json"

type Request struct {
	Namespace string          `json:"namespace"`
	Payload   json.RawMessage `json:"payload"`
}

func (b *Request) Unmarshal(data []byte) error {
	return json.Unmarshal(data, b)
}
