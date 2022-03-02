//
//  @File : rdb.go
//	@Author : WerBen
//  @Email : 289594665@qq.com
//	@Time : 2021/2/5 11:23
//	@Desc : TODO ...
//

package redis

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

var mu sync.Mutex
var ins *Pool

type Opt struct {
	Context  context.Context
	PoolSize int
	HostPort string //redis:6379
	Password string
}

func (opt *Opt) Build() []OptItem {
	return []OptItem{
		OptHostPort(opt.HostPort),
		OptPwd(opt.Password),
		OptPoolSize(opt.PoolSize),
		OptCtx(opt.Context),
	}
}

type Pool struct {
	opt        *Opt
	redisArray [16][]*Redis
	muNewRedis sync.Mutex
}

type IOptItem interface {
	apply(*Opt)
}

type OptItem struct {
	inject func(opt *Opt)
}

func (item *OptItem) apply(opt *Opt) {
	item.inject(opt)
}

func NewOptItem(inject func(opt *Opt)) OptItem {
	return OptItem{
		inject: inject,
	}
}

func OptCtx(ctx context.Context) OptItem {
	return NewOptItem(func(opt *Opt) {
		opt.Context = ctx
	})
}

func OptHostPort(hostPort string) OptItem {
	return NewOptItem(func(opt *Opt) {
		opt.HostPort = hostPort
	})
}

func OptPwd(pwd string) OptItem {
	return NewOptItem(func(opt *Opt) {
		opt.Password = pwd
	})
}

func OptPoolSize(size int) OptItem {
	return NewOptItem(func(opt *Opt) {
		opt.PoolSize = size
	})
}

func Db(index int) *Redis {
	if nil == ins {
		log.Fatalf("Error the rdb is not initialized \n")
		return nil
	}
	return ins.Get(index)
}

func NewInstance(opt *Opt) func() {
	ins = &Pool{
		opt: opt,
	}
	for k := range ins.redisArray {
		ins.redisArray[k] = make([]*Redis, opt.PoolSize)
	}
	return ins.Destroy
}

func Init(opts ...OptItem) func() {
	mu.Lock()
	defer mu.Unlock()

	// default options
	opt := &Opt{
		Context:  context.Background(),
		PoolSize: 1,
		HostPort: "127.0.0.1:6379",
	}

	// set options by args
	for _, o := range opts {
		o.apply(opt)
	}
	fmt.Printf("init redis pool %s success\n", opt.HostPort)
	return NewInstance(opt)
}

func (p *Pool) Destroy() {
	for i, v := range p.redisArray {
		for j, k := range v {
			if nil != k {
				k.Destroy()
				p.redisArray[i][j] = nil
			}
		}
	}
}

func (p *Pool) Get(db int) *Redis {
	if db < 0 || db > 15 {
		log.Fatalf("Error the db index is out of bounds\n")
		return nil
	}
	length := len(p.redisArray[db])

	index := 0
	if 1 < length {
		rand.Seed(time.Now().UnixNano())
		index = rand.Intn(length)
	}
	if nil == p.redisArray[db][index] {
		p.muNewRedis.Lock()
		defer p.muNewRedis.Unlock()
		p.redisArray[db][index] = NewRedis(
			ROptCtx(p.opt.Context),
			ROptHostPort(p.opt.HostPort),
			ROptDb(db),
			ROptPwd(p.opt.Password))
	}
	return p.redisArray[db][index]
}
