//
//  @File : my.go
//	@Author : WerBen
//  @Email : 289594665@qq.com
//	@Time : 2021/2/22 19:16
//	@Desc : TODO ...
//

package mysql

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Opt struct {
	Context context.Context
	User    string
	Pwd     string
	Host    string
	Port    string
	Db      string

	//最大空闲连接数,默认为0，表示不使用空闲连接池，
	//即一个连接如果不使用，不会放入空闲连接池。
	//因此，这种方式不会复用连接，每次执行SQL语句，都会重新建立新的连接
	MaxIdle int
	//最大连接数, 默认为0，不限制连接数
	MaxOpen int
	//连接最久能存活多少小时
	MaxLifeHour int
}

func (opt *Opt) Build() []OptItem {

	items := []OptItem{
		Ctx(opt.Context),
		Host(opt.Host),
		Port(opt.Port),
		User(opt.User),
		Pwd(opt.Pwd),
		Database(opt.Db),
	}

	if opt.MaxIdle > 0 {
		items = append(items, MaxIdle(opt.MaxIdle))
	}
	if opt.MaxOpen > 0 {
		items = append(items, MaxOpen(opt.MaxOpen))
	}
	if opt.MaxLifeHour > 0 {
		items = append(items, MaxLifeHour(opt.MaxLifeHour))
	}
	return items
}

type OptItem func(opt *Opt)

func Ctx(ctx context.Context) OptItem {
	return func(opt *Opt) {
		opt.Context = ctx
	}
}

func MaxIdle(maxIdle int) OptItem {
	return func(opt *Opt) {
		opt.MaxIdle = maxIdle
	}
}

func MaxOpen(maxOpen int) OptItem {
	return func(opt *Opt) {
		opt.MaxOpen = maxOpen
	}
}

func MaxLifeHour(maxLifeHour int) OptItem {
	return func(opt *Opt) {
		opt.MaxLifeHour = maxLifeHour
	}
}

func Host(host string) OptItem {
	return func(opt *Opt) {
		opt.Host = host
	}
}

func User(user string) OptItem {
	return func(opt *Opt) {
		opt.User = user
	}
}

func Pwd(pwd string) OptItem {
	return func(opt *Opt) {
		opt.Pwd = pwd
	}
}

func Port(port string) OptItem {
	return func(opt *Opt) {
		opt.Port = port
	}
}

func Database(database string) OptItem {
	return func(opt *Opt) {
		opt.Db = database
	}
}

type My struct {
	gorm.DB
	Opt *Opt
}

func NewMy(opts ...OptItem) *My {
	opt := &Opt{
		Host:        "127.0.0.1",
		Port:        "3306",
		User:        "root",
		Pwd:         "root",
		Db:          "mysql",
		MaxOpen:     100,
		MaxIdle:     1,
		MaxLifeHour: 1,
	}
	// set options by args
	for _, o := range opts {
		o(opt)
	}

	//dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := opt.User + ":" + opt.Pwd + "@tcp(" +
		opt.Host + ":" + opt.Port + ")/" + opt.Db +
		"?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if nil == db || nil != err {
		log.Fatalf("Error the my db connect error \n")
	}

	sqlDB, err := db.DB()
	if nil == sqlDB || nil != err {
		log.Fatalf("Error the my sql.db connect error \n")
	}
	sqlDB.SetMaxIdleConns(opt.MaxIdle)
	sqlDB.SetMaxOpenConns(opt.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Duration(opt.MaxLifeHour) * time.Hour)

	fmt.Printf("connect mysql %s success\n", opt.Host+":"+opt.Port)
	my := &My{
		*db,
		opt,
	}
	return my
}

var mu sync.Mutex
var instance *My

func Init(opts ...OptItem) {
	mu.Lock()
	defer mu.Unlock()

	instance = NewMy(opts...)
}

func Db() *gorm.DB {
	if nil == instance {
		log.Fatalf("Error the my db is not initialized \n")
		return nil
	}
	return &instance.DB
}
