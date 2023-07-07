// Copyright 2020 dongfg
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gocronexpr

import (
	"fmt"
	"github.com/willf/bitset"
	"strconv"
	"strings"
	"time"
)

// CronExpr is parse result with no exported fields
type CronExpr struct {
	expression string
	location   *time.Location

	months      *bitset.BitSet
	daysOfMonth *bitset.BitSet
	daysOfWeek  *bitset.BitSet
	hours       *bitset.BitSet
	minutes     *bitset.BitSet
	seconds     *bitset.BitSet
}

// ScheduleOptions by cron expr
type ScheduleOptions struct {
	Start  *time.Time
	End    *time.Time
	Finish func()
}

type calendar struct {
	year  int
	month int
	day   int
	hour  int
	min   int
	sec   int
	nsec  int
	loc   *time.Location

	time time.Time
}

const (
	constYear       = 0
	constMonth      = 1
	constDayOfMonth = 2
	constDayOfWeek  = 3
	constHourOfDay  = 4
	constMinute     = 5
	constSecond     = 6
)

// New cron expr, return error if parse fail
func New(expression string, location *time.Location) (*CronExpr, error) {
	c := &CronExpr{
		expression: expression,
		location:   location,

		months:      bitset.New(12),
		daysOfMonth: bitset.New(31),
		daysOfWeek:  bitset.New(7),
		hours:       bitset.New(24),
		minutes:     bitset.New(60),
		seconds:     bitset.New(60),
	}

	if err := c.parse(); err != nil {
		return c, err
	}
	return c, nil
}

// Next time calculated based on the given time.
func (c *CronExpr) Next(t *time.Time) (time.Time, error) {
	var base time.Time
	if t == nil {
		base = time.Now()
	} else {
		base = *t
	}
	cal := newCalendar(base, c.location)
	originalTimestamp := cal.time.Unix()
	if err := c.doNext(cal, cal.year); err != nil {
		return time.Time{}, err
	}
	if cal.time.Unix() == originalTimestamp {
		cal.add(constSecond, 1)
		if err := c.doNext(cal, cal.year); err != nil {
			return time.Time{}, err
		}
	}

	return cal.time, nil
}

// Run function periodically by cron expr
func (c *CronExpr) Run(fn func(), options *ScheduleOptions) {
	base := time.Now()
	for {
		next, err := c.Next(&base)
		if err != nil {
			fmt.Println("error get next run time", err)
			return
		}

		// stop after end time
		if options.End != nil && next.After(*options.End) {
			break
		}
		// start after start time
		if options.Start != nil && next.After(*options.Start) {
			next = *options.Start
		}
		<-time.After(next.Sub(base))
		fn()
		base = next
	}
	if options.Finish != nil {
		options.Finish()
	}
}

func (c *CronExpr) doNext(cal *calendar, dot int) error {
	var resets []int

	second := cal.sec
	var emptyList []int
	updateSecond := c.findNext(c.seconds, second, cal, constSecond, constMinute, emptyList)
	if second == updateSecond {
		resets = append(resets, constSecond)
	}

	minute := cal.min
	updateMinute := c.findNext(c.minutes, minute, cal, constMinute, constHourOfDay, resets)
	if minute == updateMinute {
		resets = append(resets, constMinute)
	} else {
		if err := c.doNext(cal, dot); err != nil {
			return err
		}
	}

	hour := cal.hour
	updateHour := c.findNext(c.hours, hour, cal, constHourOfDay, constDayOfWeek, resets)
	if hour == updateHour {
		resets = append(resets, constHourOfDay)
	} else {
		if err := c.doNext(cal, dot); err != nil {
			return err
		}
	}

	dayOfWeek := cal.getDayOfWeek()
	dayOfMonth := cal.day
	updateDayOfMonth, err := c.findNextDay(cal, c.daysOfMonth, dayOfMonth, c.daysOfWeek, dayOfWeek, resets)
	if err != nil {
		return err
	}
	if dayOfMonth == updateDayOfMonth {
		resets = append(resets, constDayOfMonth)
	} else {
		if err := c.doNext(cal, dot); err != nil {
			return err
		}
	}

	month := cal.month
	updateMonth := c.findNext(c.months, month, cal, constMonth, constYear, resets)
	if month != updateMonth {
		if cal.year-dot > 4 {
			return fmt.Errorf("invalid cron expression \"%s\" led to runaway search for next trigger", c.expression)
		}
		if err := c.doNext(cal, dot); err != nil {
			return err
		}
	}

	return nil
}

func (c *CronExpr) findNextDay(cal *calendar, daysOfMonth *bitset.BitSet, dayOfMonth int, daysOfWeek *bitset.BitSet, dayOfWeek int, resets []int) (int, error) {
	count := 0
	max := 366
	// the DAY_OF_WEEK values in java.util.Calendar start with 1 (Sunday),
	// but in the cron pattern, they start with 0, so we subtract 1 here
	for ((!daysOfMonth.Test(uint(dayOfMonth))) || !daysOfWeek.Test(uint(dayOfWeek-1))) && (count < max) {
		cal.add(constDayOfMonth, 1)
		dayOfMonth = cal.day
		dayOfWeek = cal.getDayOfWeek()
		cal.reset(resets)
		count++
	}
	if count >= max {
		return dayOfMonth, fmt.Errorf("overflow in day for expression %s", c.expression)
	}
	return dayOfMonth, nil
}

func (c *CronExpr) findNext(bits *bitset.BitSet, value int, cal *calendar, field int, nextField int, lowerOrders []int) int {
	nextValue, has := bits.NextSet(uint(value))
	if !has {
		cal.add(nextField, 1)
		cal.reset([]int{field})
		nextValue, _ = bits.NextSet(0)
	}
	if nextValue != uint(value) {
		cal.set(field, int(nextValue))
		cal.reset(lowerOrders)
	}

	return int(nextValue)
}

func (c *CronExpr) parse() error {
	fields := strings.Fields(c.expression)
	if len(fields) != 6 {
		return fmt.Errorf("cron expression must consist of 6 fields (found %d in \"%s\")", len(fields), c.expression)
	}

	if err := c.setNumberHits(c.seconds, fields[0], 0, 60); err != nil {
		return err
	}
	if err := c.setNumberHits(c.minutes, fields[1], 0, 60); err != nil {
		return err
	}
	if err := c.setNumberHits(c.hours, fields[2], 0, 24); err != nil {
		return err
	}
	if err := c.setDaysOfMonth(c.daysOfMonth, fields[3]); err != nil {
		return err
	}
	if err := c.setMonths(c.months, fields[4]); err != nil {
		return err
	}
	if err := c.setDays(c.daysOfWeek, replaceOrdinals(fields[5], "SUN,MON,TUE,WED,THU,FRI,SAT"), 8); err != nil {
		return err
	}

	if c.daysOfWeek.Test(7) {
		// Sunday can be represented as 0 or 7
		c.daysOfWeek.Set(0)
		c.daysOfWeek.Clear(7)
	}

	return nil
}

func replaceOrdinals(value string, commaSeparatedList string) string {
	list := strings.Split(commaSeparatedList, ",")
	for i := 0; i < len(list); i++ {
		item := strings.ToUpper(list[i])
		value = strings.ReplaceAll(strings.ToUpper(value), item, strconv.Itoa(i))
	}
	return value
}

func (c *CronExpr) setDaysOfMonth(bits *bitset.BitSet, field string) error {
	max := 31
	if err := c.setDays(bits, field, max+1); err != nil {
		return err
	}
	bits.Clear(0)
	return nil
}

func (c *CronExpr) setDays(bits *bitset.BitSet, field string, max int) error {
	if strings.Contains(field, "?") {
		field = "*"
	}
	return c.setNumberHits(bits, field, 0, max)
}

func (c *CronExpr) setMonths(bits *bitset.BitSet, value string) error {
	max := 12
	value = replaceOrdinals(value, "FOO,JAN,FEB,MAR,APR,MAY,JUN,JUL,AUG,SEP,OCT,NOV,DEC")
	months := bitset.New(13)
	if err := c.setNumberHits(months, value, 1, max+1); err != nil {
		return err
	}
	for i := 1; i <= max; i++ {
		if months.Test(uint(i)) {
			bits.Set(uint(i - 1))
		}
	}
	return nil
}

func (c *CronExpr) setNumberHits(bits *bitset.BitSet, value string, min int, max int) error {
	fields := strings.Split(value, ",")
	for _, field := range fields {
		if !strings.Contains(field, "/") {
			r, err := c.getRange(field, min, max)
			if err != nil {
				return err
			}
			setRange(bits, r[0], r[1]+1)
		} else {
			split := strings.Split(field, "/")
			if len(split) > 2 {
				return fmt.Errorf("incrementer has more than two fields: '%s' in expression \"%s\"", field, c.expression)
			}
			r, err := c.getRange(split[0], min, max)
			if err != nil {
				return err
			}
			if !strings.Contains(split[0], "-") {
				r[1] = max - 1
			}
			delta, err := strconv.Atoi(split[1])
			if err != nil {
				return err
			}
			if delta <= 0 {
				return fmt.Errorf("incrementer delta must be 1 or higher: '%s' in expression \"%s\"", field, c.expression)
			}
			for i := r[0]; i <= r[1]; i += delta {
				bits.Set(uint(i))
			}
		}
	}
	return nil
}

func (c *CronExpr) getRange(field string, min int, max int) ([]int, error) {
	var result = make([]int, 2)
	if strings.Contains(field, "*") {
		result[0] = min
		result[1] = max - 1
		return result, nil
	}
	if !strings.Contains(field, "-") {
		n, err := strconv.Atoi(field)
		if err != nil {
			return result, err
		}
		result[0], result[1] = n, n
	} else {
		split := strings.Split(field, "-")
		if len(split) > 2 {
			return result, fmt.Errorf("range has more than two fields: '%s' in expression \"%s\"", field, c.expression)
		}
		n1, err := strconv.Atoi(split[0])
		if err != nil {
			return result, err
		}
		n2, err := strconv.Atoi(split[1])
		if err != nil {
			return result, err
		}
		result[0], result[1] = n1, n2
	}
	if result[0] >= max || result[1] >= max {
		return result, fmt.Errorf("range exceeds maximum (%d): '%s' in expression \"%s\"", max, field, c.expression)
	}
	if result[0] < min || result[1] < min {
		return result, fmt.Errorf("range less than minimum (%d): '%s' in expression \"%s\"", max, field, c.expression)
	}
	if result[0] > result[1] {
		return result, fmt.Errorf("invalid inverted range (%d): '%s' in expression \"%s\"", max, field, c.expression)
	}
	return result, nil
}

func setRange(b *bitset.BitSet, from int, to int) {
	if from == to {
		return
	}
	for i := from; i < to; i++ {
		b.Set(uint(i))
	}
}

func newCalendar(t time.Time, location *time.Location) *calendar {
	cal := &calendar{
		loc:   location,
		year:  t.Year(),
		month: int(t.Month()) - 1,
		day:   t.Day(),
		hour:  t.Hour(),
		min:   t.Minute(),
		sec:   t.Second(),
		nsec:  0,
	}

	cal.time = time.Date(cal.year, time.Month(cal.month+1), cal.day, cal.hour, cal.min, cal.sec, cal.nsec, cal.loc)
	return cal
}

func (cal *calendar) add(field int, amount int) {
	switch field {
	case constYear:
		cal.year = cal.year + amount
	case constMonth:
		cal.month = cal.month + amount
	case constDayOfMonth:
		cal.day = cal.day + amount
	case constHourOfDay:
		cal.hour = cal.hour + amount
	case constMinute:
		cal.min = cal.min + amount
	case constSecond:
		cal.sec = cal.sec + amount
	case constDayOfWeek:
		cal.day = cal.day + amount
	}
	cal.align()
}

func (cal *calendar) set(field int, value int) {
	switch field {
	case constYear:
		cal.year = value
	case constMonth:
		cal.month = value
	case constDayOfMonth:
		cal.day = value
	case constHourOfDay:
		cal.hour = value
	case constMinute:
		cal.min = value
	case constSecond:
		cal.sec = value
	}
	cal.align()
}

func (cal *calendar) reset(fields []int) {
	for _, field := range fields {
		if field == constDayOfMonth {
			cal.set(field, 1)
		} else {
			cal.set(field, 0)
		}
	}
	cal.align()
}

func (cal *calendar) align() {
	t := time.Date(cal.year, time.Month(cal.month+1), cal.day, cal.hour, cal.min, cal.sec, cal.nsec, cal.loc)
	cal.year = t.Year()
	cal.month = int(t.Month()) - 1
	cal.day = t.Day()
	cal.hour = t.Hour()
	cal.min = t.Minute()
	cal.sec = t.Second()
	cal.time = t
}

// start with 1 (Sunday)
func (cal *calendar) getDayOfWeek() int {
	return int(cal.time.Weekday()) + 1
}
