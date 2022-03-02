package action

import (
	"git.aimore.com/golang/json"
	"git.aimore.com/golang/websocket"
)

type ActionReceiver struct {
}

func (h *ActionReceiver) Error(id string, request *Request, err error) error {
	return h.Resp(id, &Response{
		Namespace: request.Namespace,
		Payload: map[string]interface{}{
			"error": err.Error(),
		},
	})
}

func (h *ActionReceiver) Resp(id string, response *Response) error {
	if response != nil {
		payload, _ := json.Marshal(response)
		return websocket.Send(id, payload)
	}
	return nil
}

func (h *ActionReceiver) Recv(id string, payload []byte) error {
	request := new(Request)
	if err := request.Unmarshal(payload); err != nil {
		return h.Error(id, request, err)
	}
	response, err := RunAction(id, request)
	if err != nil {
		return h.Error(id, request, err)
	}
	return h.Resp(id, response)
}
