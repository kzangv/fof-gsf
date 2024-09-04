package cache

import (
	"github.com/kzangv/gsf-fof/cron"
	"github.com/kzangv/gsf-fof/cron/schedule"
)

const (
	NegativeTimeoutPrefix  = "t:"
	PositiveDurationPrefix = "d:"
)

type PullHandle func() (interface{}, error)

type NewScheduleFunc func(int) schedule.Interface

func DefaultNewSchedule(sec int) schedule.Interface {
	return schedule.NewDelaySchedule(sec)
}

type Cache struct {
	cron   *cron.Cron
	handle NewScheduleFunc
}

func (c *Cache) Init(cron *cron.Cron, handle NewScheduleFunc) {
	if handle == nil {
		handle = DefaultNewSchedule
	}
	c.cron, c.handle = cron, handle
}
