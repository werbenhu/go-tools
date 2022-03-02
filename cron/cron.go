package cron

import (
	"fmt"

	"github.com/rs/xid"
	"github.com/werbenhu/go-tools/cron/rc"
)

var c *rc.Cron
var idmap *BiMap //双向哈希：bidirectional map

func init() {
	idmap = NewBiMap()
	c = rc.New(rc.WithParser(rc.NewParser(
		rc.SecondOptional | rc.Minute | rc.Hour | rc.Dom | rc.Month | rc.Dow | rc.Descriptor,
	)))
}

func StartById(taskId string, p *Format, fn func(string)) (string, error) {

	//如果任务已经存在，直接返回当前任务ID
	if _, ok := idmap.Get(taskId); ok {
		return taskId, nil
	}
	//定时任务回调
	callback := func(cid rc.EntryID) {
		if tid, ok := idmap.GetInverse(cid); ok {
			fn(tid.(string))
		} else {
			fmt.Printf("cron cronId:%d not exist\n", cid)
		}
	}
	//启动定时任务
	cronId, err := c.AddFunc(p.Parse(), callback)
	if err != nil {
		fmt.Printf("start cron err:%s\n", err)
	}

	idmap.Insert(taskId, cronId)
	return taskId, err
}

func StartNative(format string, fn func(string)) (string, error) {
	//cron内部的id是int， 每次重新启动都是从0开始的
	//这里要给他匹配一个全局唯一的ID
	cronId, err := c.AddFunc(format, func(cid rc.EntryID) {
		if tid, ok := idmap.GetInverse(cid); ok {
			fn(tid.(string))
		} else {
			fmt.Printf("cron cid:%d not exist\n", cid)
		}
	})
	if err != nil {
		fmt.Printf("start cron err:%s\n", err)
	}

	taskId := xid.New().String()
	idmap.Insert(taskId, cronId)
	return taskId, err
}

func Start(p *Format, fn func(string)) (string, error) {
	//cron内部的id是int， 每次重新启动都是从0开始的
	//这里要给他匹配一个全局唯一的ID
	cronId, err := c.AddFunc(p.Parse(), func(cid rc.EntryID) {
		if tid, ok := idmap.GetInverse(cid); ok {
			fn(tid.(string))
		} else {
			fmt.Printf("cron cid:%d not exist\n", cid)
		}
	})
	if err != nil {
		fmt.Printf("start cron err:%s\n", err)
	}

	taskId := xid.New().String()
	idmap.Insert(taskId, cronId)
	return taskId, err
}

func Cancel(taskId string) {
	if cid, ok := idmap.Get(taskId); ok {
		idmap.Delete(taskId)
		c.Remove(cid.(rc.EntryID))
	} else {
		fmt.Printf("cron taskId:%s not exist\n", taskId)
	}
}

func Count() int {
	return len(c.Entries())
}

func Run() {
	c.Start()
}
