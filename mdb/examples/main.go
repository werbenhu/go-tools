package main

import (
	"context"
	"fmt"

	"git.aimore.com/golang/mdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func initMongo() {
	ctx := context.Background()
	opt := mdb.Opt{
		Context: ctx,
		Url:     "mongodb://root:bu12345@183.134.21.93:27017",
	}
	mdb.Init(opt.Build()...)
}

type Action struct {
	Did  string
	Act  string
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

	coll := mdb.Collection("bugDb", "dev")

	opts := options.FindOne()
	var result Dev
	filter := bson.D{{"openid", "Dy42uXrxcWkj1"}}
	err := coll.FindOne(context.TODO(), filter, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Println("record does not exist")
	}
	fmt.Printf("result:%+v\n", result)
}
