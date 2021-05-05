package gocronexpr

import (
	"errors"
	"fmt"
	"github.com/willf/bitset"
	"strconv"
	"strings"
	"time"
)

type cronexpr struct {
	expression  string
	months      *bitset.BitSet
	daysOfMonth *bitset.BitSet
	daysOfWeek  *bitset.BitSet
	hours       *bitset.BitSet
	minutes     *bitset.BitSet
	seconds     *bitset.BitSet
}

type calendar struct {
	year  int
	month time.Month
	day   int
	hour  int
	min   int
	sec   int
	nsec  int
	loc   *time.Location
}

var timeField = map[string]int{
	
}

// Next time calculated based on the given time.
func Next(expression string, base time.Time) (time.Time, error) {
	cronexpr := &cronexpr{
		expression: expression,
	}
	if err := cronexpr.parse(); err != nil {
		return base, errors.New("cron expression must consist of 6 fields")
	}
	return time.Now(), nil
}

func (cronexpr *cronexpr) parse() error {
	fields := strings.Split(cronexpr.expression, " ")
	if len(fields) != 6 {
		return errors.New(fmt.Sprintf("cron expression must consist of 6 fields (found %d in \"%s\")", len(fields), cronexpr))
	}

	if err := cronexpr.setNumberHits(cronexpr.seconds, fields[0], 0, 60); err != nil {
		return err
	}
	if err := cronexpr.setNumberHits(cronexpr.minutes, fields[1], 0, 60); err != nil {
		return err
	}
	if err := cronexpr.setNumberHits(cronexpr.hours, fields[2], 0, 24); err != nil {
		return err
	}
	if err := cronexpr.setDaysOfMonth(cronexpr.daysOfMonth, fields[3]); err != nil {
		return err
	}
	if err := cronexpr.setMonths(cronexpr.months, fields[4]); err != nil {
		return err
	}
	if err := cronexpr.setDays(cronexpr.daysOfWeek, replaceOrdinals(fields[5], "SUN,MON,TUE,WED,THU,FRI,SAT"), 8); err != nil {
		return err
	}

	if cronexpr.daysOfWeek.Test(7) {
		// Sunday can be represented as 0 or 7
		cronexpr.daysOfWeek.Set(0)
		cronexpr.daysOfWeek.Clear(7)
	}

	return nil
}

func (cronexpr *cronexpr) next(base time.Time) (time.Time, error) {
	/*calendar := calendar{
		year:  base.Year(),
		month: base.Month(),
		day:   base.Day(),
		hour:  base.Hour(),
		min:   base.Minute(),
		sec:   base.Second(),
		nsec:  0,
		loc:   time.Local,
	}

	originalTimestamp := base.Unix()*/

	return time.Now(), nil
}

func doNext(calendar *calendar, dot int) {
	// var rests []int

	// second := calendar.sec
}

func (cronexpr *cronexpr) findNext(bits *bitset.BitSet, value int, calendar *calendar, field int, nextField int, lowerOrders []int) int {
	nextValue, has := bits.NextSet(uint(value))
	if !has {

	}
}

func (cronexpr *cronexpr) setDaysOfMonth(bits *bitset.BitSet, field string) error {
	max := 31
	if err := cronexpr.setDays(bits, field, max+1); err != nil {
		return err
	}
	bits.Set(0)
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

func (cronexpr *cronexpr) setDays(bits *bitset.BitSet, field string, max int) error {
	if strings.Contains(field, "?") {
		field = "*"
	}
	return cronexpr.setNumberHits(bits, field, 0, max) //fixme 这边需要传指针吗
}

func (cronexpr *cronexpr) setMonths(bits *bitset.BitSet, value string) error {
	max := 12
	value = replaceOrdinals(value, "FOO,JAN,FEB,MAR,APR,MAY,JUN,JUL,AUG,SEP,OCT,NOV,DEC")
	months := bitset.New(13)
	if err := cronexpr.setNumberHits(months, value, 1, max+1); err != nil {
		return err
	}
	for i := 1; i <= max; i++ {
		if months.Test(uint(i)) {
			bits.Set(uint(i - 1))
		}
	}
	return nil
}

func (cronexpr *cronexpr) setNumberHits(bits *bitset.BitSet, value string, min int, max int) error {
	fields := strings.Split(value, ",")
	for _, field := range fields {
		if !strings.Contains(field, "/") {
			r, err := cronexpr.getRange(field, min, max)
			if err != nil {
				return err
			}
			setRange(bits, r[0], r[1]+1)
		} else {
			split := strings.Split(field, "/")
			if len(split) > 2 {
				return errors.New(fmt.Sprintf("incrementer has more than two fields: '%s' in expression \"%s\"", field, cronexpr.expression))
			}
			r, err := cronexpr.getRange(split[0], min, max)
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
				return errors.New(fmt.Sprintf("incrementer delta must be 1 or higher: '%s' in expression \"%s\"", field, cronexpr.expression))
			}
			for i := r[0]; i <= r[1]; i += delta {
				bits.Set(uint(i))
			}
		}
	}
	return nil
}

func (cronexpr *cronexpr) getRange(field string, min int, max int) ([]int, error) {
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
			return result, errors.New(fmt.Sprintf("range has more than two fields: '%s' in expression \"%s\"", field, cronexpr.expression))
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
		return result, errors.New(fmt.Sprintf("range exceeds maximum (%d): '%s' in expression \"%s\"", max, field, cronexpr.expression))
	}
	if result[0] < min || result[1] < min {
		return result, errors.New(fmt.Sprintf("range less than minimum (%d): '%s' in expression \"%s\"", max, field, cronexpr.expression))
	}
	if result[0] > result[1] {
		return result, errors.New(fmt.Sprintf("invalid inverted range (%d): '%s' in expression \"%s\"", max, field, cronexpr.expression))
	}
	return result, nil
}

func setRange(b *bitset.BitSet, from int, to int) {
	if from == to {
		return
	}
	for i := from; i <= to; i++ {
		b.Set(uint(i))
	}
}
