package main

import (
	"context"
	"fmt"
	"time"

	"git.aimore.com/golang/mqtt"
)

func initTlsMqtt() {
	ctx := context.Background()
	mqtt.Init(mqtt.Opt{
		Ctx:      ctx,
		Host:     "127.0.0.1",
		Port:     "1884",
		ClientId: "device-srv",
		IsTls:    true,
		CaFile:   "certs/root.crt",
		CertFile: "certs/client.crt",
		KeyFile:  "certs/client.key",
	})
}

func initMqtt() {
	ctx := context.Background()
	mqtt.Init(mqtt.Opt{
		Ctx:      ctx,
		Host:     "127.0.0.1",
		Port:     "1884",
		ClientId: "device-srv",
	})
}

type Data struct {
	Code int
	Msg  string
}

func mqttHandler(topic string, data []byte) {
	fmt.Printf("mqttHandler msg:%s\n", string(data))
}

func runSubs() {
	go mqtt.Sub("/test", mqttHandler)
}

func runPubs() {
	time.Sleep(time.Second)
	mqtt.PubEx("/test", &Data{
		Code: 101,
		Msg:  "success",
	})
	mqtt.Pub("/test", "123456")
	mqtt.Pub("/test", []byte("abcdeft"))
}

func main() {
	exitCh := make(chan error)

	// initMqtt()
	initTlsMqtt()
	runSubs()
	runPubs()

	<-exitCh
}
