//
//  @File : nsq.go
//	@Author : werben
//  @Email : werben@qq.com
//	@Time : 2021/2/4 17:34
//	@Desc : TODO ...
//

package nsq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"

	"github.com/nsqio/go-nsq"
)

func fatalf(format string, a ...interface{}) {
	debug.PrintStack()
	log.Fatalf("cause: %s\n", fmt.Sprintf(format, a...))
}

type Handler func(payload []byte) error

type DefaultHandler struct {
	H Handler
}

func (handler *DefaultHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		return nil
	}
	return handler.H(m.Body)
}

type Opt struct {
	Context    context.Context
	LookupHost string
	NsqHost    string
}

type OptItem func(opt *Opt)

func (opt *Opt) Build() []OptItem {
	return []OptItem{
		OptCtx(opt.Context),
		OptNsqHost(opt.NsqHost),
		OptLookupHost(opt.LookupHost),
	}
}

func OptCtx(context context.Context) OptItem {
	return func(opt *Opt) {
		opt.Context = context
	}
}

func OptNsqHost(host string) OptItem {
	return func(opt *Opt) {
		opt.NsqHost = host
	}
}

func OptLookupHost(lookupHost string) OptItem {
	return func(opt *Opt) {
		opt.LookupHost = lookupHost
	}
}

type Nsq struct {
	Opt             *Opt
	Config          *nsq.Config
	Producer        *nsq.Producer
	ConsumerContext context.Context
	ConsumerStop    context.CancelFunc
}

type Destroy func()

var mu sync.Mutex
var n *Nsq

func Init(opts ...OptItem) Destroy {
	mu.Lock()
	defer mu.Unlock()
	if nil == n {
		// default options
		opt := &Opt{
			Context:    context.Background(),
			NsqHost:    "127.0.0.1:4150",
			LookupHost: "",
		}

		// set options by args
		for _, o := range opts {
			o(opt)
		}
		n = &Nsq{Opt: opt}
		n.Config = nsq.NewConfig()
		n.ConsumerContext, n.ConsumerStop = context.WithCancel(n.Opt.Context)

		pro, err := nsq.NewProducer(n.Opt.NsqHost, n.Config)
		//logger := log.New(os.Stderr, "", log.Flags())
		//pro.SetLogger(logger, nsq.LogLevelError)
		n.Producer = pro
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("connect nsqd %s success\n", opt.NsqHost)
	}
	return n.Destroy
}

func (n *Nsq) Destroy() {
	n.Producer.Stop()
	n.ConsumerStop()
}

func (n *Nsq) Public(topic string, payload []byte) {
	if nil != n.Producer {
		err := n.Producer.Publish(topic, payload)
		if err != nil {
			log.Println(err)
		}
	}
}

func (n *Nsq) Subscribe(topic string, channel string, handler Handler) {

	consumer, err := nsq.NewConsumer(topic, channel, n.Config)

	if consumer == nil || err != nil {
		fatalf("Error topic:%s, channel:%s, err: %s\n", topic, channel, err)
	}

	logger := log.New(os.Stderr, "", log.Flags())
	consumer.SetLogger(logger, nsq.LogLevelError)
	defer consumer.Stop()

	consumer.AddHandler(&DefaultHandler{
		H: handler,
	})

	if n.Opt.LookupHost != "" {
		err = consumer.ConnectToNSQLookupd(n.Opt.LookupHost)
	} else {
		err = consumer.ConnectToNSQD(n.Opt.NsqHost)
	}

	if err != nil {
		fatalf("Error: %s\n", err)
	}

	<-n.ConsumerContext.Done()
}

func Produce(topic string, payload interface{}) {
	if nil == n {
		fatalf("Error: nsq not initialized!!!")
	}

	switch body := payload.(type) {
	case string:
		n.Public(topic, []byte(body))
	case []byte:
		n.Public(topic, body)
	default:
		fatalf("Produce()'s parameter payload's type should be [ string, []byte ], please use ProduceEx() instead")
	}
}

func ProduceEx(topic string, payload interface{}) {
	if nil == n {
		fatalf("Error: nsq not initialized!!!")
	}
	switch payload.(type) {
	case string, []byte:
		fatalf("ProduceEx()'s parameter payload's type can't be [ string, []byte ], please use Produce() instead")
	default:
		body, _ := json.Marshal(payload)
		n.Public(topic, body)
	}
}

func Consume(topic string, channel string, handler Handler) {
	if nil == n {
		fatalf("Error: nsq not initialized!!!")
	}
	n.Subscribe(topic, channel, handler)
	fmt.Printf("consume done")
}
