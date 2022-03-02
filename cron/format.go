package cron

import (
	"errors"
	"strconv"
)

type Format struct {
	items    []Item
	timezone string
}

func NewFormat(tz string) *Format {
	p := new(Format)
	p.items = make([]Item, 6)
	p.timezone = tz
	return p
}

func (p *Format) Timezone(tz string) {
	p.timezone = tz
}

func (p *Format) Second(s string) error {
	val, err := strconv.Atoi(s)
	if val >= 60 || val < 0 {
		return errors.New("Err: second number can't bigger then 60")
	}

	if err != nil {
		return err
	}
	p.items[0].add(val)
	return nil
}

func (p *Format) SecondLoop(gap string) error {
	p.items[0].loop(gap)
	return nil
}

func (p *Format) Minute(m string) error {
	val, err := strconv.Atoi(m)
	if val >= 60 || val < 0 {
		return errors.New("Err: second number can't bigger then 60 and smaller then 0")
	}

	if err != nil {
		return err
	}
	p.items[1].add(val)
	return nil
}

func (p *Format) MinuteLoop(gap string) error {
	p.items[1].loop(gap)
	return nil
}

func (p *Format) Hour(h string) error {
	val, err := strconv.Atoi(h)
	if val >= 24 || val < 0 {
		return errors.New("Err: second number can't bigger then 23 and smaller then 0")
	}

	if err != nil {
		return err
	}

	p.items[2].add(val)
	return nil
}

func (p *Format) HourLoop(gap string) error {
	p.items[2].loop(gap)
	return nil
}

func (p *Format) Day(d string) error {
	val, err := strconv.Atoi(d)
	if val >= 31 {
		return errors.New("Err: second number can't bigger then 31 and smaller then 0")
	}

	if err != nil {
		return err
	}
	p.items[3].add(val)
	return nil
}

func (p *Format) DayLoop(gap string) error {
	p.items[3].loop(gap)
	return nil
}

func (p *Format) Week(w string) error {
	val, err := strconv.Atoi(w)
	if val >= 7 {
		return errors.New("Err: second number can't bigger then 6  and smaller then 0")
	}

	if err != nil {
		return err
	}
	p.items[5].add(val)
	return nil
}

func (p *Format) Parse() string {
	ret := ""

	if p.timezone != "" {
		ret = "CRON_TZ=" + p.timezone + " "
	}

	for i, item := range p.items {
		ret += item.parse()
		if i < 5 {
			ret += " "
		}
	}
	return ret
}
