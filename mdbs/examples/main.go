package main

import (
	"context"
	"fmt"
	"reflect"

	"github.com/werbenhu/go-tools/mdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func initMongo() {
	ctx := context.Background()
	opt := mdb.Opt{
		Context: ctx,
		Url:     "mongodb://dayuaimore:TslUfWIc65DbrtrFMQezfGuvgAx3VVm@cn.iot.thedayu.com:27019",
	}
	mdb.Init(opt.Build()...)
}

type Item struct {
	Name string
}

type Action struct {
	Did  string
	Act  interface{}
	Type string
}

type List struct {
	Name   string
	Src    string
	Time   string
	Room   string
	Action []Action
}

type Dev struct {
	Openid string
	List   List
}

func main() {
	initMongo()

	//参考文档: https://github.com/mongodb/mongo-go-driver
	//https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo#pkg-examples

	coll := mdb.Collection("werben", "action")
	action := &Action{
		Did: "aaa",
		Act: map[string]interface{}{
			"name": "123",
		},
	}

	_, err := coll.InsertOne(context.TODO(), action)
	if err != nil {
		fmt.Printf("Device Insert err:%s\n", err)
	}
	if err != nil {
		fmt.Printf("err:%s\n", err)
	}

	fopts := options.FindOne()
	filter := bson.D{{"did", "werben"}}

	// bson.MarshalExtJSON()

	var a Action
	err = coll.FindOne(context.TODO(), filter, fopts).Decode(&a)
	if err == mongo.ErrNoDocuments {
		fmt.Printf("Device Get :%+v, record does not exist\n", filter)
	}

	fmt.Printf("a:%+v\n", &a)
	fmt.Printf("action:%+v\n", action)

	fmt.Printf("eq:%t\n", reflect.DeepEqual(&a, action))

	var e interface{}
	var f interface{}

	e = int32(10)
	f = int(10)
	fmt.Printf("eq:%t\n", e.(int32) == f)

}
