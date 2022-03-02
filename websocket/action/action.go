package action

import (
	"errors"
	"fmt"
	"sync"
)

var actions []Action
var mu sync.Mutex

var Recevier ActionReceiver

type Action interface {
	Namespace() string
	Excute(userId string, req *Request) (*Response, error)
}

func ActionRegister(action Action) {
	mu.Lock()
	defer mu.Unlock()
	actions = append(actions, action)
}

func RunAction(userId string, request *Request) (*Response, error) {
	for _, action := range actions {
		if action.Namespace() == request.Namespace ||
			action.Namespace() == "*" {
			return action.Excute(userId, request)
		}
	}
	errStr := fmt.Sprintf("websocket namespace: %s not exist", request.Namespace)
	return nil, errors.New(errStr)
}
