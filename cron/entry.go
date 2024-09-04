package cron

import (
	"gitee.com/kzangv/gsf-fof/cron/schedule"
	"time"
)

type Job interface {
	Init()
	Run(time.Time)
	Destroy()
}

type ScheduleJob interface {
	schedule.Interface
	Job
}

type ScheduleRun func(t time.Time)

type WrapJob ScheduleRun

func (f WrapJob) Init()           {}
func (f WrapJob) Run(t time.Time) { f(t) }
func (f WrapJob) Destroy()        {}

type WrapScheduleJob struct {
	Job
	schedule.Interface
}

type entry struct {
	Name     string
	Schedule ScheduleJob
	Next     time.Time
	Prev     time.Time
}

type entries []*entry

func (pq entries) Len() int {
	return len(pq)
}

func (pq entries) Less(i, j int) bool {
	return pq[i].Next.Before(pq[j].Next)
}

func (pq entries) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *entries) Push(x interface{}) {
	item := x.(*entry)
	*pq = append(*pq, item)
}

func (pq *entries) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}
