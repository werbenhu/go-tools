package action

import "git.aimore.com/golang/json"

type Response struct {
	Namespace string      `json:"namespace"`
	Payload   interface{} `json:"payload"`
}

func (r *Response) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Response) SetNamespace(namespace string) *Response {
	r.Namespace = namespace
	return r
}

func (r *Response) SetPayload(payload interface{}) *Response {
	r.Payload = payload
	return r
}
