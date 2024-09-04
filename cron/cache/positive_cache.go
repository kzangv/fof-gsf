package cache

import (
	"fmt"
	"gitee.com/kzangv/gsf-fof/cron/schedule"
	"time"
)

type _DurationItem struct {
	val interface{}
	get PullHandle
	err error
	sch schedule.Interface
}

func (c *_DurationItem) Init()                       {}
func (c *_DurationItem) Run(time.Time)               {}
func (c *_DurationItem) Destroy()                    {}
func (c *_DurationItem) Value() (interface{}, error) { return c.val, c.err }
func (c *_DurationItem) Next(t time.Time) *time.Time {
	c.val, c.err = c.get()
	return c.sch.Next(t)
}

type RefreshCache struct {
	Cache
}

func (c *RefreshCache) Add(name string, sec int, get PullHandle) {
	e := &_DurationItem{
		get: get,
		sch: c.handle(sec),
	}
	c.cron.AddScheduleJob(PositiveDurationPrefix+name, e)
}
func (c *RefreshCache) Get(name string) (interface{}, error) {
	sj := c.cron.Job(PositiveDurationPrefix + name)
	if sj == nil {
		return nil, fmt.Errorf("key [%s] is no find", name)
	}
	j := sj.(*_DurationItem)
	return j.Value()
}
