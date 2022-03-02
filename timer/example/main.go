package main

import (
	"fmt"
	"time"

	"git.aimore.com/golang/timer"
)

func main() {
	timer.Run()
	timer.Start(5, func(id string) {
		fmt.Printf("after 5 sec id:%s\n", id)
		fmt.Printf("count:%d\n", timer.Count())
	})

	timer.Start(7, func(id string) {
		fmt.Printf("after 52 sec id:%s\n", id)
		fmt.Printf("count:%d\n", timer.Count())
	})

	scheduleId, _ := timer.Schedule(3, func(id string) {
		fmt.Printf("schedule sec id:%s\n", id)
		fmt.Printf("schedule count:%d\n", timer.Count())
	})
	time.Sleep(8 * time.Second)
	timer.Cancel(scheduleId)

	fmt.Printf("count:%d\n", timer.Count())
	time.Sleep(12 * time.Second)
	fmt.Printf("count:%d\n", timer.Count())
}
