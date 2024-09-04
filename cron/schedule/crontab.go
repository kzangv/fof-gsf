package schedule

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	starBit = 1 << 63
)

type crontab struct {
	Second, Minute, Hour, Dom, Month, Dow uint64
}

func (s *crontab) dayMatches(t time.Time) bool {
	var (
		domMatch = 1<<uint(t.Day())&s.Dom > 0
		dowMatch = 1<<uint(t.Weekday())&s.Dow > 0
	)

	if s.Dom&starBit > 0 || s.Dow&starBit > 0 {
		return domMatch && dowMatch
	}
	return domMatch || dowMatch
}

func (s *crontab) Next(t time.Time) *time.Time {
	t = t.Add(1*time.Second - time.Duration(t.Nanosecond())*time.Nanosecond)
	added, yearLimit := false, t.Year()+5
WRAP:
	if t.Year() > yearLimit {
		return nil
	}

	for 1<<uint(t.Month())&s.Month == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 1, 0)

		if t.Month() == time.January {
			goto WRAP
		}
	}

	for !s.dayMatches(t) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 0, 1)

		if t.Day() == 1 {
			goto WRAP
		}
	}

	for 1<<uint(t.Hour())&s.Hour == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
		}
		t = t.Add(1 * time.Hour)

		if t.Hour() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Minute())&s.Minute == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
		}
		t = t.Add(1 * time.Minute)

		if t.Minute() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Second())&s.Second == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
		}
		t = t.Add(1 * time.Second)

		if t.Second() == 0 {
			goto WRAP
		}
	}

	return &t
}

func NewCrontabSchedule(spec string) (Interface, error) {
	return crontabParser(spec)
}

var (
	seconds = bounds{0, 59, nil}
	minutes = bounds{0, 59, nil}
	hours   = bounds{0, 23, nil}
	dom     = bounds{1, 31, nil}
	months  = bounds{1, 12, map[string]uint{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}}
	dow = bounds{0, 6, map[string]uint{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}}
)

type bounds struct {
	min, max uint
	names    map[string]uint
}

func parseIntOrName(expr string, names map[string]uint) (uint, error) {
	if names != nil {
		if namedInt, ok := names[strings.ToLower(expr)]; ok {
			return namedInt, nil
		}
	}
	return mustParseInt(expr)
}

func mustParseInt(expr string) (uint, error) {
	num, err := strconv.Atoi(expr)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse int from %s: %s", expr, err)
	}
	if num < 0 {
		return 0, fmt.Errorf("Negative number (%d) not allowed: %s", num, expr)
	}
	return uint(num), nil
}

func getBits(min, max, step uint) uint64 {
	var bits uint64
	if step == 1 {
		return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}
	for i := min; i <= max; i += step {
		bits |= 1 << i
	}
	return bits
}

func getField(field string, r *bounds) (uint64, error) {
	var bits uint64
	ranges := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	for _, expr := range ranges {
		if v, err := getRange(expr, r); err != nil {
			return 0, err
		} else {
			bits |= v
		}
	}
	return bits, nil
}

func getRange(expr string, r *bounds) (uint64, error) {
	var (
		start, end, step uint
		rangeAndStep     = strings.Split(expr, "/")
		lowAndHigh       = strings.Split(rangeAndStep[0], "-")
		singleDigit      = len(lowAndHigh) == 1
	)

	var extra uint64
	var err error
	if lowAndHigh[0] == "*" || lowAndHigh[0] == "?" {
		start = r.min
		end = r.max
		extra = starBit
	} else {
		start, err = parseIntOrName(lowAndHigh[0], r.names)
		if err != nil {
			return 0, err
		}
		switch len(lowAndHigh) {
		case 1:
			end = start
		case 2:
			end, err = parseIntOrName(lowAndHigh[1], r.names)
			if err != nil {
				return 0, err
			}
		default:
			return 0, fmt.Errorf("Too many hyphens: %s", expr)
		}
	}

	switch len(rangeAndStep) {
	case 1:
		step = 1
	case 2:
		step, err = mustParseInt(rangeAndStep[1])
		if err != nil {
			return 0, err
		}
		if singleDigit {
			end = r.max
		}
		if step > 1 {
			extra = 0
		}
	default:
		return 0, fmt.Errorf("Too many slashes: %s", expr)
	}

	if start < r.min {
		return 0, fmt.Errorf("Beginning of range (%d) below minimum (%d): %s", start, r.min, expr)
	}
	if end > r.max {
		return 0, fmt.Errorf("End of range (%d) above maximum (%d): %s", end, r.max, expr)
	}
	if start > end {
		return 0, fmt.Errorf("Beginning of range (%d) beyond end of range (%d): %s", start, end, expr)
	}
	if step == 0 {
		return 0, fmt.Errorf("step of range should be a positive number: %s", expr)
	}
	return getBits(start, end, step) | extra, nil
}

func crontabParser(spec string) (*crontab, error) {
	fields := strings.Fields(spec)
	if len(fields) != 5 && len(fields) != 6 {
		return nil, fmt.Errorf("Expected 5 or 6 fields, found %d: %s", len(fields), spec)
	}

	if len(fields) == 5 {
		fields = append(fields, "*")
	}
	schedule := &crontab{}

	data := []struct {
		fields string
		bounds *bounds
		value  *uint64
	}{
		{fields[0], &seconds, &schedule.Second},
		{fields[1], &minutes, &schedule.Minute},
		{fields[2], &hours, &schedule.Hour},
		{fields[3], &dom, &schedule.Dom},
		{fields[4], &months, &schedule.Month},
		{fields[5], &dow, &schedule.Dow},
	}
	for k := range data {
		if v, err := getField(data[k].fields, data[k].bounds); err != nil {
			return nil, err
		} else {
			*(data[k].value) = v
		}
	}
	return schedule, nil
}
