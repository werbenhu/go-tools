package main

import (
	"context"
	"fmt"
	"time"

	"github.com/werbenhu/go-tools/nsq"
)

func initNsq() {
	ctx := context.Background()
	opt := nsq.Opt{
		Context: ctx,
		// LookupHost: "218.91.230.204:4150",
		NsqHost: "218.91.230.204:4150",
	}
	nsq.Init(opt.Build()...)
}

type Data struct {
	Code int
	Msg  string
}

func nsqHandler(payload []byte) error {
	fmt.Printf("nsqHandler payload:%s\n", string(payload))
	return nil
}

func runProduce() {
	time.Sleep(time.Second)
	nsq.Produce("test", "test")
	nsq.ProduceEx("test", &Data{
		Code: 101,
		Msg:  "success",
	})
}

func runConsume() {
	go nsq.Consume("test", "testCh", nsqHandler)
}

func main() {
	exitCh := make(chan error)

	initNsq()
	runConsume()
	runProduce()

	<-exitCh
}
