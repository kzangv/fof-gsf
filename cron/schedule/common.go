package schedule

import "time"

type Interface interface {
	Next(time.Time) *time.Time
}
