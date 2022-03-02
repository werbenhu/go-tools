package mqtts

import (
	"context"
	"sync"
)

var mu sync.Mutex
var brokers map[string]*Mqtt

type Opt struct {
	Ctx   context.Context
	Items map[string]MOpt `json:"items"`
}

func Init(opt Opt) func() {
	mu.Lock()
	defer mu.Unlock()

	brokers = make(map[string]*Mqtt)
	destroys := make([]func(), 0)
	var destroy func()

	for k, o := range opt.Items {
		o.Ctx = opt.Ctx
		brokers[k], destroy = NewMqtt(o)
		destroys = append(destroys, destroy)
	}
	return func() {
		for _, destroy := range destroys {
			destroy()
		}
	}
}

func Map() map[string]*Mqtt {
	return brokers
}

func Get(key string) *Mqtt {
	if client, ok := brokers[key]; ok {
		return client
	}
	return nil
}
