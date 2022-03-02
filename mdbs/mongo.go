//
//  @File : my.go
//	@Author : WerBen
//  @Email : 289594665@qq.com
//	@Time : 2021/2/22 19:16
//	@Desc : TODO ...
//

package mdbs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MOpt struct {
	Context context.Context
	//mongodb://username:password@127.0.0.01:27017
	Url string

	MinIdle     uint64
	MaxIdle     uint64
	MaxLifeHour uint64
	IsTls       bool
	CaFile      string
}

func (opt *MOpt) Build() []OptItem {

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

type OptItem func(opt *MOpt)

func Ctx(ctx context.Context) OptItem {
	return func(opt *MOpt) {
		opt.Context = ctx
	}
}

func MaxIdle(maxIdle uint64) OptItem {
	return func(opt *MOpt) {
		opt.MaxIdle = maxIdle
	}
}

func MinIdle(minIdle uint64) OptItem {
	return func(opt *MOpt) {
		opt.MinIdle = minIdle
	}
}

func MaxLifeHour(maxLifeHour uint64) OptItem {
	return func(opt *MOpt) {
		opt.MaxLifeHour = maxLifeHour
	}
}

func IsTls(isTls bool) OptItem {
	return func(opt *MOpt) {
		opt.IsTls = isTls
	}
}

func CaFile(cafile string) OptItem {
	return func(opt *MOpt) {
		opt.CaFile = cafile
	}
}

func Url(url string) OptItem {
	return func(opt *MOpt) {
		opt.Url = url
	}
}

type Mdb struct {
	mongo.Client
	Opt *MOpt
	Ctx context.Context
}

func (m *Mdb) Destroy() {
	m.Client.Disconnect(m.Ctx)
}

func (m *Mdb) Collection(db string, collection string) *mongo.Collection {
	return m.Client.Database(db).Collection(collection)
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
	opt := &MOpt{
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
		log.Fatalf("Error mongodb failed to ping server \n")
	}

	fmt.Printf("connect mongodb %s success\n", opt.Url)
	return &Mdb{*client, opt, ctx}
}
