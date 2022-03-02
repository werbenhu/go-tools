package main

import (
	"context"
	"fmt"

	"github.com/werbenhu/go-tools/redis"
)

func initRedis() {
	ctx := context.Background()
	opt := redis.Opt{
		Context:  ctx,
		PoolSize: 1,
		HostPort: "183.134.21.92:16379",
		Password: "&^%aks!c@Nm8^.sa-`50)wq*e,}e&^%!ds5&54aa!!sd+~",
	}
	redis.Init(opt.Build()...)
}

func main() {
	initRedis()

	rdb := redis.Db(15)
	// rdb.SetWithTime("werben", "huang", time.Minute)
	// ret, err := rdb.Get("werben")
	// fmt.Printf("ret1:%s, err:%s\n", ret, err)

	// rdb.Delete("werben")
	// ret, err = rdb.Get("werben")
	// fmt.Printf("ret2:%s, err:%s\n", ret, err)

	key := "zwerben"
	rdb.ZAdd(key, 100, "a")
	rdb.ZAdd(key, 200, "b")
	rdb.ZAdd(key, 300, "c")
	rdb.ZAdd(key, 0.0005, "a")

	ret, err := rdb.ZScore(key, "a")

	if ret == 0 {
		fmt.Printf("nto exist\n")
	}

	fmt.Printf("%f, %s\n", ret, err)
	ret, err = rdb.ZScore(key, "e")
	fmt.Printf("%f, %s\n", ret, err)
}
