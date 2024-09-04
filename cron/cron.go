package cron

import (
	"container/heap"
	"gitee.com/kzangv/gsf-fof/cron/schedule"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ScheduleEventAdd    = 1
	ScheduleEventChange = 2
	ScheduleEventRemove = 3
	ScheduleEventResize = 4
)

type event struct {
	tm    time.Time
	entry *entry
	event int
}

type Resize interface {
	Check(maxCap, len int) bool
	NewCap(len int) int
}

type Cron struct {
	dataLock sync.RWMutex
	dataMap  map[string]*entry

	running uint32
	resize  Resize
	ticker  *time.Ticker
	change  chan event
	stop    chan struct{}
}

func (s *Cron) AddScheduleJob(name string, cmd ScheduleJob) {
	now := time.Now()
	if next := cmd.Next(now); next != nil {
		if cmd != nil {
			cmd.Init()
		}
		s.dataLock.RLock()
		e, ok := s.dataMap[name]
		s.dataLock.RUnlock()

		if ok {
			e.Schedule, e.Next = cmd, *next
			s.change <- event{
				event: ScheduleEventChange,
				tm:    now,
				entry: e,
			}
		} else {
			v := &entry{
				Name:     name,
				Schedule: cmd,
				Next:     *next,
			}

			s.dataLock.Lock()
			s.dataMap[name] = v
			s.dataLock.Unlock()

			s.change <- event{
				event: ScheduleEventAdd,
				tm:    now,
				entry: v,
			}
		}
	} else {
		s.Remove(name)
	}
}

func (s *Cron) AddJob(name string, sch schedule.Interface, cmd Job) {
	s.AddScheduleJob(name, &WrapScheduleJob{cmd, sch})
}
func (s *Cron) AddFunc(name string, sch schedule.Interface, cmd ScheduleRun) {
	s.AddJob(name, sch, WrapJob(cmd))
}

func (s *Cron) Remove(name string) {
	s.dataLock.RLock()
	e, ok := s.dataMap[name]
	s.dataLock.RUnlock()

	if ok {
		s.dataLock.Lock()
		delete(s.dataMap, name)
		s.dataLock.Unlock()

		e.Schedule.Destroy()
		s.change <- event{
			event: ScheduleEventRemove,
			tm:    time.Now(),
			entry: e,
		}
	}
}

func (s *Cron) Entries() entries {
	s.dataLock.RLock()
	defer s.dataLock.RUnlock()

	ret := make(entries, 0, 10)
	for _, e := range s.dataMap {
		ret = append(ret, &entry{
			Next:     e.Next,
			Prev:     e.Prev,
			Schedule: e.Schedule,
			Name:     e.Name,
		})
	}
	return ret
}

func (s *Cron) Entry(name string) *entry {
	s.dataLock.RLock()
	defer s.dataLock.RUnlock()

	if e, ok := s.dataMap[name]; ok {
		return &entry{
			Next:     e.Next,
			Prev:     e.Prev,
			Schedule: e.Schedule,
			Name:     e.Name,
		}
	}
	return nil
}

func (s *Cron) Job(name string) ScheduleJob {
	s.dataLock.RLock()
	defer s.dataLock.RUnlock()

	if v, ok := s.dataMap[name]; ok {
		return v.Schedule
	}
	return nil
}

func (s *Cron) resetTick(queue *entries, now time.Time) {
	if (*queue).Len() > 0 {
		front := (*queue)[0]
		if now.Before(front.Next) {
			s.ticker.Reset(front.Next.Sub(now))
		} else {
			s.ticker.Reset(time.Nanosecond)
		}
	} else {
		s.ticker.Reset(time.Hour * 24)
	}
}

func (s *Cron) runSchedule(queue *entries, now time.Time) {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	for queue.Len() > 0 {
		front := (*queue)[0]
		if now.Before(front.Next) {
			break
		}
		e := heap.Pop(queue).(*entry)
		if _, ok := s.dataMap[e.Name]; ok {
			next := e.Schedule.Next(now)
			go e.Schedule.Run(now)
			if next == nil {
				e.Schedule.Destroy()
				delete(s.dataMap, e.Name)
			} else {
				e.Prev, e.Next = e.Next, *next
				heap.Push(queue, e)
			}
		}
	}
}

func (s *Cron) Resize() {
	s.change <- event{
		event: ScheduleEventResize,
		tm:    time.Now(),
		entry: nil,
	}
}

func (s *Cron) ResizeMap() {
	s.dataLock.Lock()
	dataMap := make(map[string]*entry, len(s.dataMap)+100)
	for k, v := range s.dataMap {
		dataMap[k] = v
	}
	s.dataMap = dataMap
	s.dataLock.Unlock()
}

func (s *Cron) Start() {
	s.dataLock.Lock()
	if s.dataMap == nil {
		s.dataMap = make(map[string]*entry, 100)
	}
	s.dataLock.Unlock()
	if atomic.CompareAndSwapUint32(&s.running, 0, 1) {
		s.ticker = time.NewTicker(time.Hour * 24)
		go func() {
			defer func() { s.stop <- struct{}{} }()

			// 初始化 queue
			queue := make(entries, 0, 10)
			s.resetTick(&queue, time.Now())

			for {
				select {
				case now := <-s.ticker.C:
					s.runSchedule(&queue, now)
					s.resetTick(&queue, time.Now())
				case ev := <-s.change:
					switch ev.event {
					case ScheduleEventAdd:
						heap.Push(&queue, ev.entry)
						s.resetTick(&queue, ev.tm)
					case ScheduleEventChange:
						for k := range queue {
							if queue[k].Name == ev.entry.Name {
								heap.Fix(&queue, k)
							}
						}
						s.resetTick(&queue, ev.tm)
					case ScheduleEventRemove:
						// noting to do
					case ScheduleEventResize:
						if s.resize != nil {
							if s.resize.Check(cap(queue), len(queue)) {
								newQueue := make(entries, 0, s.resize.NewCap(len(queue)))
								for k := range queue {
									newQueue = append(newQueue, queue[k])
								}
								queue = newQueue
							}
						}
					}
				case <-s.stop:
					return
				}
			}
		}()
	}
}

func (s *Cron) Stop() {
	if atomic.CompareAndSwapUint32(&s.running, 1, 0) {
		s.dataLock.RLock()
		defer s.dataLock.RUnlock()

		defer close(s.stop)
		defer close(s.change)
		defer s.ticker.Stop()

		s.stop <- struct{}{}
		<-s.stop
		wg := sync.WaitGroup{}
		for _, value := range s.dataMap {
			wg.Add(1)
			go func() {
				defer wg.Done()
				j := value.Schedule
				if j != nil {
					j.Destroy()
				}
			}()
		}
		wg.Wait()
	}
}

func NewCron(h Resize, chgBuff int) *Cron {
	return &Cron{
		running: 0,
		ticker:  nil,
		change:  make(chan event, chgBuff),
		stop:    make(chan struct{}),
		resize:  h,
	}
}
