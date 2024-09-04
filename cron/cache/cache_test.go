package cache

import (
	"gitee.com/kzangv/gsf-fof/cron"
	"sync"
	"testing"
	"time"
)

func GetUser() (interface{}, error) {
	return 10, nil
}

func Change(wg *sync.WaitGroup, sec int, f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Second * time.Duration(sec))
		f()
	}()
}

func TestPositiveCache(t *testing.T) {
	c := cron.NewCron(nil, 0)
	c.Start()
	wg := sync.WaitGroup{}
	t.Logf("%s: begin", time.Now().Format("15-04-05"))

	ca := RefreshCache{}
	ca.Init(c, nil)

	ca.Add("u", 5, GetUser)

	Change(&wg, 3, func() {
		v, _ := ca.Get("u")
		t.Logf("%s: %v", time.Now().Format("15-04-05"), v)
	})

	Change(&wg, 15, func() {
		v, _ := ca.Get("u")
		t.Logf("%s: %v", time.Now().Format("15-04-05"), v)
	})

	time.Sleep(time.Second * 20)
	wg.Wait()
}

func TestNegativeCache(t *testing.T) {
	c := cron.NewCron(nil, 0)
	c.Start()
	wg := sync.WaitGroup{}
	t.Logf("%s: begin", time.Now().Format("15-04-05"))

	ca := TimeoutCache{}
	ca.Init(c, nil)

	ca.Add("u", 5, GetUser)

	Change(&wg, 3, func() {
		v, _ := ca.Get("u")
		t.Logf("%s: %v", time.Now().Format("15-04-05"), v)
	})

	Change(&wg, 15, func() {
		v, _ := ca.Get("u")
		t.Logf("%s: %v", time.Now().Format("15-04-05"), v)
	})

	time.Sleep(time.Second * 20)
	wg.Wait()
}
