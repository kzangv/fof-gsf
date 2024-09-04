package cron

import (
	"fmt"
	"github.com/kzangv/gsf-fof/cron/schedule"
	"sync"
	"testing"
	"time"
)

func GetPrintFunc(key string, ts *testing.T) ScheduleRun {
	cnt := 1
	return func(t time.Time) {
		ts.Logf("%s %s[cnt:%d]", key, t.Format("15:04:05"), cnt)
		cnt++
	}
}

func AddDelayPrintJob(sec int, t *testing.T, c *Cron) string {
	key := fmt.Sprintf("print_%d", sec)
	c.AddFunc(key, schedule.NewDelaySchedule(sec), GetPrintFunc(key, t))
	return key
}

func AddDelayLimitPrintJob(sec, limit int, t *testing.T, c *Cron) string {
	key := fmt.Sprintf("print_%d", sec)
	c.AddFunc(key, schedule.NewLimitSchedule(uint(limit), schedule.NewDelaySchedule(sec)), GetPrintFunc(key, t))
	return key
}

func Change(wg *sync.WaitGroup, sec int, f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Second * time.Duration(sec))
		f()
	}()
}

func TestDelay(t *testing.T) {
	c := NewCron(nil, 0)
	c.Start()
	wg := sync.WaitGroup{}
	t.Logf("%s: begin", time.Now().Format("15-04-05"))

	k1 := AddDelayPrintJob(3, t, c)
	_ = AddDelayPrintJob(6, t, c)
	_ = AddDelayPrintJob(9, t, c)
	Change(&wg, 4, func() {
		c.AddFunc(k1, schedule.NewDelaySchedule(12), GetPrintFunc(k1, t))
	})

	time.Sleep(time.Second * 30)
	wg.Wait()
}

func TestLimit(t *testing.T) {
	c := NewCron(nil, 0)
	c.Start()
	wg := sync.WaitGroup{}
	t.Logf("%s: begin", time.Now().Format("15-04-05"))
	_ = AddDelayLimitPrintJob(3, 3, t, c)
	time.Sleep(time.Second * 30)
	wg.Wait()
}
