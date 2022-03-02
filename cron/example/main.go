package main

import (
	"fmt"
	"time"

	"github.com/werbenhu/go-tools/cron"
)

func main() {
	cron.Run()
	format := cron.NewFormat("Asia/Shanghai")
	// gap := format.Correct("2021-12-07 21:27:05")
	// format.Second("00")
	// format.Minute("17")
	// format.Hour("14")
	format.SecondLoop("1")

	id, _ := cron.StartById("c6o6lpkfe3hhih1todg1", format, func(id string) {
		fmt.Printf("cron cb id:%s\n", id)
		fmt.Printf("cb:%d\n", cron.Count())
	})

	// id, _ := cron.Start(format, func(id string) {
	// 	fmt.Printf("cron cb id:%s\n", id)
	// 	fmt.Printf("cb:%d\n", cron.Count())
	// })
	fmt.Printf("cron id:%s\n", id)

	time.Sleep(5 * time.Second)

	cron.Cancel(id)
	fmt.Printf("c:%d\n", cron.Count())

	select {}
}
