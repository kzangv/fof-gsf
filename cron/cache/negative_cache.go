package cache

import (
	"fmt"
	"time"
)

type _TimeoutItem struct {
	_DurationItem
	use bool
}

func (c *_TimeoutItem) Next(t time.Time) *time.Time {
	if c.use {
		c.use = false
		c.val, c.err = c.get()
		return c.sch.Next(t)
	}
	return nil
}
func (c *_TimeoutItem) setUse() {
	c.use = true
}

type TimeoutCache struct {
	Cache
}

func (c *TimeoutCache) Add(name string, sec int, get PullHandle) {
	e := &_TimeoutItem{
		_DurationItem: _DurationItem{
			get: get,
			sch: c.handle(sec),
		},
		use: true,
	}
	c.cron.AddScheduleJob(NegativeTimeoutPrefix+name, e)
}
func (c *TimeoutCache) Get(name string) (interface{}, error) {
	sj := c.cron.Job(NegativeTimeoutPrefix + name)
	if sj == nil {
		return nil, fmt.Errorf("key [%s] is no find", name)
	}
	j := sj.(*_TimeoutItem)
	j.setUse()
	return j.Value()
}
