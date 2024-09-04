package schedule

import (
	"time"
)

type limit struct {
	limit uint
	sch   Interface
}

func (s *limit) Next(t time.Time) *time.Time {
	if s.limit > 0 {
		s.limit--
		return s.sch.Next(t)
	}
	return nil
}

func NewLimitSchedule(cnt uint, sch Interface) Interface {
	return &limit{
		limit: cnt,
		sch:   sch,
	}
}
