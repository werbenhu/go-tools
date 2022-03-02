package main

import (
	"context"
	"fmt"

	"git.aimore.com/golang/mysql"
)

func initMysql() {
	ctx := context.Background()
	opt := mysql.Opt{
		Context: ctx,
		User:    "root",
		Pwd:     "aimore123456",
		Host:    "218.91.230.204",
		Port:    "23306",
		Db:      "test",
	}
	mysql.Init(opt.Build()...)
}

type Teacher struct {
	Id       int
	Name     string
	Age      int
	ClazzId  int
	CreateAt int64
}

func main() {
	initMysql()

	//参考gorm v2.X文档: https://gorm.io/zh_CN/docs/
	var teacher Teacher
	mysql.Db().Where("name = ?", "wenwu").Find(&teacher)
	fmt.Printf("teahcer:%+v\n", teacher)
}
