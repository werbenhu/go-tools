package timer

import "time"

type DefaultScheduler struct {
	Interval time.Duration
}

func (s *DefaultScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}
