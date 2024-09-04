package schedule

import (
	"time"
)

type delay struct {
	delay time.Duration
}

func (s *delay) Next(t time.Time) *time.Time {
	v := t.Add(s.delay - time.Duration(t.Nanosecond())*time.Nanosecond)
	return &v
}

func NewDelaySchedule(sec int) Interface {
	duration := time.Second * time.Duration(sec)
	return &delay{
		delay: duration - time.Duration(duration.Nanoseconds())%time.Second,
	}
}
