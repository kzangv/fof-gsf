package cron

const (
	ResizeOverTime = 3
)

type CommonResize struct {
	overTime,
	CheckTime uint8

	TickerQueueIdleLimit int
}

func (r *CommonResize) Check(c, l int) bool {
	ret, fSize := false, c-l
	limit, cSize := r.TickerQueueIdleLimit, 100
	if fSize > limit {
		limit = limit / 3
		if limit < cSize {
			limit = cSize
		}
		if limit < fSize {
			ret = true
		}
	}
	if ret {
		r.overTime, ret = r.overTime+1, false
		if r.overTime > ResizeOverTime {
			ret, r.overTime = true, 0
		}
	} else if r.overTime > 0 {
		r.overTime = 0
	}
	return ret
}

func (r *CommonResize) NewCap(l int) int {
	cl := r.TickerQueueIdleLimit
	cl = cl / 5
	if cl < 10 {
		cl = 10
	} else if cl > 200 {
		cl = 200
	}
	return l + cl
}

func NewCommonResize(gap uint8) *CommonResize {
	return &CommonResize{
		TickerQueueIdleLimit: 1000,
		CheckTime:            gap,
	}
}
