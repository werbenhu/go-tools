//
//  @File : redis.go.go
//	@Author : WerBen
//  @Email : 289594665@qq.com
//	@Time : 2021/2/4 20:33
//	@Desc : TODO ...
//

package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	rdb "github.com/go-redis/redis/v8"
)

type ROpt struct {
	Ctx      context.Context
	Db       int
	HostPort string //redis:6379
	Password string
}

type ROptItem func(opt *ROpt)

func ROptCtx(ctx context.Context) ROptItem {
	return func(opt *ROpt) {
		opt.Ctx = ctx
	}
}

func ROptHostPort(hostPort string) ROptItem {
	return func(opt *ROpt) {
		opt.HostPort = hostPort
	}
}

func ROptPwd(pwd string) ROptItem {
	return func(opt *ROpt) {
		opt.Password = pwd
	}
}

func ROptDb(db int) ROptItem {
	return func(opt *ROpt) {
		opt.Db = db
	}
}

//Redis redis object
type Redis struct {
	cli    *rdb.Client
	config *rdb.Options
	opt    *ROpt
}

type Destroy func()

//New make new redis obj
func NewRedis(opts ...ROptItem) *Redis {

	r := new(Redis)
	// default options
	opt := &ROpt{
		Ctx:      context.Background(),
		Db:       0,
		HostPort: "127.0.0.1:6379",
	}

	// set options by args
	for _, o := range opts {
		o(opt)
	}
	r.opt = opt
	r.Connect()
	return r
}

//Connect connect to rdb
func (r *Redis) Connect() error {

	config := rdb.Options{
		Addr:     r.opt.HostPort,
		Password: r.opt.Password,
		DB:       r.opt.Db,
	}
	cli := rdb.NewClient(&config)
	_, err := cli.Ping(r.opt.Ctx).Result()
	if err != nil {
		log.Printf("Error redis connection failed %s\n", err)
		return err
	}
	r.cli = cli
	r.config = &config
	fmt.Printf("connect redis %s success\n", r.opt.HostPort)
	return nil
}

func (r *Redis) CheckConnection() {
	_, err := r.cli.Ping(r.opt.Ctx).Result()
	if err != nil {
		r.Connect()
	}
}

//Get get value by key
func (r *Redis) Get(key string) (string, error) {
	if r.cli == nil {
		err := errors.New("redis not connection")
		return "", err
	}
	val, err := r.cli.Get(r.opt.Ctx, key).Result()
	return val, err
}

//Get get value by key
func (r *Redis) GetBytes(key string) ([]byte, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return nil, err
	}
	val, err := r.cli.Get(r.opt.Ctx, key).Bytes()
	return val, err
}

func (r *Redis) SetWithTime(key string, val string, time time.Duration) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.Set(r.opt.Ctx, key, val, time).Result()
	return err
}

func (r *Redis) SetBytesWithTime(key string, val []byte, time time.Duration) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.Set(r.opt.Ctx, key, val, time).Result()
	return err
}

//Set set value by key
func (r *Redis) Set(key string, val string) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.Set(r.opt.Ctx, key, val, 0).Result()
	return err
}

func (r *Redis) SetBytes(key string, val []byte) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.Set(r.opt.Ctx, key, val, 0).Result()
	return err
}

func (r *Redis) SetInter(key string, val interface{}) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.Set(r.opt.Ctx, key, val, 0).Result()
	return err
}

//Keys find keys by pattern
func (r *Redis) Keys(pattern string) ([]string, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return nil, err
	}
	val, err := r.cli.Keys(r.opt.Ctx, pattern).Result()
	return val, err
}

func (r *Redis) Delete(key string) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.Del(r.opt.Ctx, key).Result()
	return err
}

func (r *Redis) Destroy() {
	if r.cli != nil {
		r.cli.Close()
	}
}

func (r *Redis) ZAdd(key string, score float64, val string) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	var Z rdb.Z
	Z.Score = score
	Z.Member = val
	_, err := r.cli.ZAdd(r.opt.Ctx, key, &Z).Result()
	return err
}

func (r *Redis) ZRange(key string, start, stop string) ([]rdb.Z, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return nil, err
	}
	var Z rdb.ZRangeBy
	Z.Min = start
	Z.Max = stop
	val, err := r.cli.ZRangeByScoreWithScores(r.opt.Ctx, key, &Z).Result()
	return val, err
}

func (r *Redis) ZDel(key string, min, max string) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}

	_, err := r.cli.ZRemRangeByScore(r.opt.Ctx, key, min, max).Result()
	return err
}

func (r *Redis) ZScore(key string, member string) (float64, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return -1, err
	}

	val, err := r.cli.ZScore(r.opt.Ctx, key, member).Result()
	return val, err
}

func (r *Redis) HSet(key string, values ...interface{}) (int64, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return 0, err
	}

	val, err := r.cli.HSet(r.opt.Ctx, key, values...).Result()
	return val, err
}

func (r *Redis) HGet(key string, filed string) (string, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return "", err
	}

	val, err := r.cli.HGet(r.opt.Ctx, key, filed).Result()
	return val, err
}

func (r *Redis) HDel(key string, filed string) (int64, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return 0, err
	}

	val, err := r.cli.HDel(r.opt.Ctx, key, filed).Result()
	return val, err
}

func (r *Redis) HKeys(key string) ([]string, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return nil, err
	}
	return r.cli.HKeys(r.opt.Ctx, key).Result()
}

func (r *Redis) HGetAll(key string) (map[string]string, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return nil, err
	}
	return r.cli.HGetAll(r.opt.Ctx, key).Result()
}

func (r *Redis) Expire(key string, expiration time.Duration) (bool, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return false, err
	}

	val, err := r.cli.Expire(r.opt.Ctx, key, expiration).Result()
	return val, err
}

func (r *Redis) LPop(key string) (string, error) {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return "", err
	}
	val, err := r.cli.LPop(r.opt.Ctx, key).Result()
	return val, err
}

func (r *Redis) LRange(key string, start int64, stop int64) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.LRange(r.opt.Ctx, key, start, stop).Result()
	return err
}

func (r *Redis) LRem(key string, start int64, stop int64) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.LRem(r.opt.Ctx, key, start, stop).Result()
	return err
}

func (r *Redis) RPush(key string, value string) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.RPush(r.opt.Ctx, key, value).Result()
	return err
}

func (r *Redis) LTrim(key string) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.LTrim(r.opt.Ctx, key, 1, -1).Result()
	return err
}

func (r *Redis) LIndex(key string, index int64) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.LIndex(r.opt.Ctx, key, index).Result()
	return err
}

func (r *Redis) LSet(key string, index int64, val string) error {
	if r.cli == nil {
		err := errors.New("RDB not connection")
		return err
	}
	_, err := r.cli.LSet(r.opt.Ctx, key, index, val).Result()
	return err
}
