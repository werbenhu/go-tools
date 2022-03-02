package mdbs

import (
	"context"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

var mu sync.Mutex
var dbs map[string]*Mdb
var opt Opt

type Opt struct {
	Context context.Context
	Area    string          `json:"area"`
	Items   map[string]MOpt `json:"items"`
}

func (o *Opt) SetArea(area string) {
	o.Area = area
}

func Init(mopt Opt) func() {
	mu.Lock()
	defer mu.Unlock()

	opt = mopt
	dbs = make(map[string]*Mdb)
	for key, item := range opt.Items {
		item.Context = opt.Context
		mdb := New(item.Build()...)
		dbs[key] = mdb
	}

	return func() {
		for _, v := range dbs {
			v.Destroy()
		}
	}
}

func Db(name string) *mongo.Client {
	if mdb, ok := dbs[name]; ok && mdb != nil {
		return &mdb.Client
	}
	log.Fatalf("Error mongodb is not initialized \n")
	return nil
}

func Get(name string) *Mdb {
	if mdb, ok := dbs[name]; ok && mdb != nil {
		return mdb
	}
	log.Fatalf("Error mongodb is not initialized \n")
	return nil
}

func GetOpt() *Opt {
	return &opt
}

func Collection(db string, collection string) *mongo.Collection {
	if mdb, ok := dbs[opt.Area]; ok && mdb != nil {
		return mdb.Collection(db, collection)
	}
	log.Fatalf("Error mongodb is not initialized \n")
	return nil
}
