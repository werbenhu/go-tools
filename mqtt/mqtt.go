//
//  @File : mqtt.go
//	@Author : werben
//  @Email : werben@qq.com
//	@Time : 2021/2/23 10:45
//	@Desc : Encapsulation of mqtt interface
//

package mqtt

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"runtime/debug"
	"sync"
	"time"

	mq "github.com/eclipse/paho.mqtt.golang"
)

var mu sync.Mutex
var ins *Mqtt

type Opt struct {
	Ctx      context.Context
	Host     string `json:"host"`
	Port     string `json:"port"`
	ClientId string `json:"clientId"`

	IsTls    bool   `json:"isTls"`
	CaFile   string `json:"caFile"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}

type Mqtt struct {
	opt            Opt
	cli            mq.Client
	handlerMu      sync.Mutex
	handlerMap     map[string][]Handler
	isReconnecting bool
	url            string
}

type Handler func(topic string, data []byte)

func NewTLSConfig(caFile string, certFile string, keyFile string) *tls.Config {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(caFile)
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic(err)
	}

	// Just to print out the client certificate..
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		panic(err)
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}
}

func NewClientOpt(m *Mqtt, opt *Opt) *mq.ClientOptions {
	options := mq.NewClientOptions().
		SetClientID(opt.ClientId).
		SetAutoReconnect(true).
		SetConnectionLostHandler(m.onConnectionLost).
		SetKeepAlive(2 * time.Second).
		SetPingTimeout(1 * time.Second).
		SetMaxReconnectInterval(time.Second).
		SetKeepAlive(time.Minute * 15).
		SetOnConnectHandler(m.onConnect).
		SetReconnectingHandler(m.onReconnecting)

	if opt.IsTls {
		tlsConfig := NewTLSConfig(opt.CaFile, opt.CertFile, opt.KeyFile)
		options.SetTLSConfig(tlsConfig)
		m.url = "mqtts://" + opt.Host + ":" + opt.Port
	} else {
		m.url = "mqtt://" + opt.Host + ":" + opt.Port
	}
	options.AddBroker(m.url)
	return options
}

func Init(opt Opt) func() {
	mu.Lock()
	defer mu.Unlock()
	ins = new(Mqtt)
	ins.handlerMap = make(map[string][]Handler)
	// mq.DEBUG = log.New(os.Stdout, "", 0)
	// mq.ERROR = log.New(os.Stdout, "", 0)

	clientOpt := NewClientOpt(ins, &opt)
	ins.opt = opt
	ins.cli = mq.NewClient(clientOpt)

	if t := ins.cli.Connect(); t.Wait() && t.Error() != nil {
		log.Printf("server:%s connect err:%s\n", ins.url, t.Error())
		panic(t.Error())
	}
	return func() {
		ins.cli.Disconnect(0)
	}
}

func (m *Mqtt) onReconnecting(cli mq.Client, opt *mq.ClientOptions) {
	log.Printf("mqtt server: %s onReconnecting\n", m.url)
	m.isReconnecting = true
}

func (m *Mqtt) onConnectionLost(cli mq.Client, err error) {
	log.Printf("mqtt server: %s onConnectionLost\n", m.url)
}

func (m *Mqtt) onConnect(cli mq.Client) {
	fmt.Printf("connect mqtt %s success\n", m.url)
	//m.handlerMu.Lock()
	//defer m.handlerMu.Unlock()
	if m.isReconnecting {
		for topic, handlers := range m.handlerMap {
			for _, handler := range handlers {
				m.Subscribe(topic, handler)
			}
		}
	}
	m.isReconnecting = false
}

func (m *Mqtt) Public(topic string, data interface{}) error {
	t := m.cli.Publish(topic, 0, false, data)
	if t.Wait(); t.Error() != nil {
		return t.Error()
	}
	return nil
}

func (m *Mqtt) Subscribe(topic string, handler Handler) {
	m.handlerMu.Lock()
	m.handlerMap[topic] = append(m.handlerMap[topic], handler)
	m.handlerMu.Unlock()

	t := m.cli.Subscribe(topic, 0, func(c mq.Client, m mq.Message) {
		go func(h Handler, t string, p []byte) {
			h(t, p)
		}(handler, m.Topic(), m.Payload())
	})

	if t.Wait(); t.Error() != nil {
		m.handlerMu.Lock()
		delete(m.handlerMap, topic)
		m.handlerMu.Unlock()
		fatalf("mqtt Subscribe err:%s\n", t.Error())
	}
}

func (m *Mqtt) Unsubscribe(topic string) {
	m.handlerMu.Lock()
	m.cli.Unsubscribe(topic)
	delete(m.handlerMap, topic)
	m.handlerMu.Unlock()
}

func fatalf(format string, a ...interface{}) {
	debug.PrintStack()
	log.Fatalf("cause: %s\n", fmt.Sprintf(format, a...))
}

func Pub(topic string, payload interface{}) {
	if nil == ins {
		fatalf("mqtt not initialized! call mqtt.Init() first!\n")
	}
	switch payload.(type) {
	case string, []byte, bytes.Buffer:
		ins.Public(topic, payload)
	default:
		fatalf("Pub()'s parameter payload's type should be [ string, []byte, bytes.Buffe ], please use PubEx() instead")
	}
}

func PubEx(topic string, payload interface{}) {
	if nil == ins {
		fatalf("mqtt not initialized! call mqtt.Init() first!\n")
	}
	switch payload.(type) {
	case string, []byte, bytes.Buffer:
		fatalf("PubEx()'s parameter payload's type can't be [ string, []byte, bytes.Buffe ], please use Pub() instead")
	default:
		data, _ := json.Marshal(payload)
		ins.Public(topic, data)
	}
}

func Sub(topic string, handler Handler) {
	if nil == ins {
		fatalf("mqtt not initialized! call mqtt.Init() first!\n")
	}
	ins.Subscribe(topic, handler)
}

func Unsub(topic string) {
	if nil == ins {
		fatalf("mqtt not initialized! call mqtt.Init() first!\n")
	}
	ins.Unsubscribe(topic)
}
