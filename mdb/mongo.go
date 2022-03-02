//
//  @File : my.go
//	@Author : WerBen
//  @Email : 289594665@qq.com
//	@Time : 2021/2/22 19:16
//	@Desc : TODO ...
//

package mdb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Opt struct {
	Context context.Context
	//mongodb://username:password@127.0.0.01:27017
	Url string

	MinIdle     uint64
	MaxIdle     uint64
	MaxLifeHour uint64
	IsTls       bool
	CaFile      string
}

func (opt *Opt) Build() []OptItem {

	items := []OptItem{
		Ctx(opt.Context),
		Url(opt.Url),
	}

	if opt.MinIdle > 0 {
		items = append(items, MinIdle(opt.MinIdle))
	}
	if opt.MaxIdle > 0 {
		items = append(items, MaxIdle(opt.MaxIdle))
	}
	if opt.MaxLifeHour > 0 {
		items = append(items, MaxLifeHour(opt.MaxLifeHour))
	}
	items = append(items, IsTls(opt.IsTls))
	items = append(items, CaFile(opt.CaFile))
	return items
}

type OptItem func(opt *Opt)

func Ctx(ctx context.Context) OptItem {
	return func(opt *Opt) {
		opt.Context = ctx
	}
}

func MaxIdle(maxIdle uint64) OptItem {
	return func(opt *Opt) {
		opt.MaxIdle = maxIdle
	}
}

func MinIdle(minIdle uint64) OptItem {
	return func(opt *Opt) {
		opt.MinIdle = minIdle
	}
}

func MaxLifeHour(maxLifeHour uint64) OptItem {
	return func(opt *Opt) {
		opt.MaxLifeHour = maxLifeHour
	}
}

func IsTls(isTls bool) OptItem {
	return func(opt *Opt) {
		opt.IsTls = isTls
	}
}

func CaFile(cafile string) OptItem {
	return func(opt *Opt) {
		opt.CaFile = cafile
	}
}

func Url(url string) OptItem {
	return func(opt *Opt) {
		opt.Url = url
	}
}

type Mdb struct {
	mongo.Client
	Opt *Opt
	Ctx context.Context
}

func (m *Mdb) Destroy() {
	m.Client.Disconnect(m.Ctx)
}

func getCustomTLSConfig(caFile string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certs, err := ioutil.ReadFile(caFile)
	if err != nil {
		return tlsConfig, err
	}
	tlsConfig.RootCAs = x509.NewCertPool()
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs)
	if !ok {
		log.Fatalf("Error mongodb Failed parsing pem file")
	}
	return tlsConfig, nil
}

func New(opts ...OptItem) *Mdb {
	opt := &Opt{
		////mongodb://username:password@127.0.0.01:27017
		Url:         "mongodb://root:123456@127.0.0.01:27017",
		MinIdle:     1,
		MaxIdle:     10,
		MaxLifeHour: 1,
		IsTls:       false,
		CaFile:      "",
	}
	// set options by args
	for _, o := range opts {
		o(opt)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOps := options.Client().
		ApplyURI(opt.Url).
		SetMaxPoolSize(opt.MaxIdle).
		SetMinPoolSize(opt.MaxIdle).
		SetMaxConnIdleTime(time.Duration(opt.MaxLifeHour) * time.Hour)

	if opt.IsTls {
		tlsConfig, err := getCustomTLSConfig(opt.CaFile)
		if err != nil {
			log.Fatalf("Error mongodb getting TLS configuration: %v \n", err)
		}
		clientOps = clientOps.SetTLSConfig(tlsConfig)
	}

	client, err := mongo.NewClient(clientOps)
	if err != nil {
		log.Fatalf("Error mongodb to create client: %v \n", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error mongodb failed to connect server %v \n", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("Error mongodb failed to ping server:%s err:%v \n", opt.Url, err)
	}

	fmt.Printf("connect mongodb %s success\n", opt.Url)
	return &Mdb{*client, opt, ctx}
}

var mu sync.Mutex
var ins *Mdb

func Init(opts ...OptItem) func() {
	mu.Lock()
	defer mu.Unlock()

	ins = New(opts...)
	return ins.Destroy
}

func Db() *mongo.Client {
	if nil == ins {
		log.Fatalf("Error mongodb is not initialized \n")
		return nil
	}
	return &ins.Client
}

func Collection(db string, collection string) *mongo.Collection {
	return ins.Client.Database(db).Collection(collection)
}
