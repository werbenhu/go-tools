package cron

import (
	"strconv"
	"strings"
)

type Item struct {
	isLoop   bool
	elems    []int
	loopText string
}

func (i *Item) parse() string {
	if i.isLoop {
		return i.loopText
	}
	return i.join("*", ",")
}

func (item *Item) loop(gap string) {
	item.isLoop = true
	item.loopText = "*/" + gap
}

func (item *Item) add(val int) {
	item.isLoop = false
	item.elems = append(item.elems, val)
}

func (item *Item) join(none string, sep string) string {
	switch len(item.elems) {
	case 0:
		return none
	case 1:
		return strconv.Itoa(item.elems[0])
	}
	n := len(sep) * (len(item.elems) - 1)
	for i := 0; i < len(item.elems); i++ {
		n += len(strconv.Itoa(item.elems[i]))
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(strconv.Itoa(item.elems[0]))
	for _, s := range item.elems[1:] {
		b.WriteString(sep)
		b.WriteString(strconv.Itoa(s))
	}
	return b.String()
}
