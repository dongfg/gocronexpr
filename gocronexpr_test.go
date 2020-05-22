package gocronexpr

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	type args struct {
		expression string
		location   *time.Location
	}
	var tests []struct {
		name    string
		args    args
		wantErr bool
	}

	validList := []string{
		"* * * 2 * *",
		"57,59 * * * * *",
		"1,3,5 * * * * *",
		"* * 4,8,12,16,20 * * *",
		"* * * * * 0-6",
		"* * * * * 0",
		"* * * * * 0",
		"* * * * 1-12 *",
		"* * * * 2 *",
		"*  *  * *  1 *",
	}
	for i, valid := range validList {
		tests = append(tests, struct {
			name    string
			args    args
			wantErr bool
		}{
			name: fmt.Sprintf("%s_%d", "valid_cron", i),
			args: args{
				expression: valid,
				location:   time.Local,
			},
			wantErr: false,
		})
	}

	invalidList := []string{
		"77 * * * * *",
		"44-77 * * * * *",
		"* 77 * * * *",
		"* 44-77 * * * *",
		"* * 27 * * *",
		"* * 23-28 * * *",
		"* * * 45 * *",
		"* * * 28-45 * *",
		"0 0 0 25 13 ?",
		"0 0 0 25 0 ?",
		"0 0 0 32 12 ?",
		"* * * * 11-13 *",
		"-5 * * * * *",
		"3-2 */5 * * * *",
		"/5 * * * * *",
		"*/0 * * * * *",
		"*/-0 * * * * *",
	}
	for i, invalid := range invalidList {
		tests = append(tests, struct {
			name    string
			args    args
			wantErr bool
		}{
			name: fmt.Sprintf("%s_%d", "invalid_cron", i),
			args: args{
				expression: invalid,
				location:   time.Local,
			},
			wantErr: true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.expression, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v, expression %v", err, tt.wantErr, tt.args.expression)
				return
			}
		})
	}
}

func Test_cronexpr_Next(t *testing.T) {
	var tests []struct {
		name       string
		expression string
		baseTime   string
		want       string
	}

	cases := [][]string{
		{"*/15 * 1-4 * * *", "2012-07-01 09:53:50", "2012-07-02 01:00:00"},
		{"*/15 * 1-4 * * *", "2012-07-01 09:53:00", "2012-07-02 01:00:00"},
		{"0 */2 1-4 * * *", "2012-07-01 09:00:00", "2012-07-02 01:00:00"},
		{"0 */2 * * * *", "2012-07-01 09:00:00", "2012-07-01 09:02:00"},
		{"0 */2 * * * *", "2013-07-01 09:00:00", "2013-07-01 09:02:00"},
		{"0 */2 * * * *", "2018-09-14 14:24:00", "2018-09-14 14:26:00"},
		{"0 */2 * * * *", "2018-09-14 14:25:00", "2018-09-14 14:26:00"},
		{"0 */20 * * * *", "2018-09-14 14:24:00", "2018-09-14 14:40:00"},
		{"* * * * * *", "2012-07-01 09:00:00", "2012-07-01 09:00:01"},
		{"* * * * * *", "2012-12-01 09:00:58", "2012-12-01 09:00:59"},
		{"10 * * * * *", "2012-12-01 09:42:09", "2012-12-01 09:42:10"},
		{"11 * * * * *", "2012-12-01 09:42:10", "2012-12-01 09:42:11"},
		{"10 * * * * *", "2012-12-01 09:42:10", "2012-12-01 09:43:10"},
		{"10-15 * * * * *", "2012-12-01 09:42:09", "2012-12-01 09:42:10"},
		{"10-15 * * * * *", "2012-12-01 21:42:14", "2012-12-01 21:42:15"},
		{"0 * * * * *", "2012-12-01 21:10:42", "2012-12-01 21:11:00"},
		{"0 * * * * *", "2012-12-01 21:11:00", "2012-12-01 21:12:00"},
		{"0 11 * * * *", "2012-12-01 21:10:42", "2012-12-01 21:11:00"},
		{"0 10 * * * *", "2012-12-01 21:11:00", "2012-12-01 22:10:00"},
		{"0 0 * * * *", "2012-09-30 11:01:00", "2012-09-30 12:00:00"},
		{"0 0 * * * *", "2012-09-30 12:00:00", "2012-09-30 13:00:00"},
		{"0 0 * * * *", "2012-09-10 23:01:00", "2012-09-11 00:00:00"},
		{"0 0 * * * *", "2012-09-11 00:00:00", "2012-09-11 01:00:00"},
		{"0 0 0 * * *", "2012-09-01 14:42:43", "2012-09-02 00:00:00"},
		{"0 0 0 * * *", "2012-09-02 00:00:00", "2012-09-03 00:00:00"},
		{"* * * 10 * *", "2012-10-09 15:12:42", "2012-10-10 00:00:00"},
		{"* * * 10 * *", "2012-10-11 15:12:42", "2012-11-10 00:00:00"},
		{"0 0 0 * * *", "2012-09-30 15:12:42", "2012-10-01 00:00:00"},
		{"0 0 0 * * *", "2012-10-01 00:00:00", "2012-10-02 00:00:00"},
		{"0 0 0 * * *", "2012-08-30 15:12:42", "2012-08-31 00:00:00"},
		{"0 0 0 * * *", "2012-08-31 00:00:00", "2012-09-01 00:00:00"},
		{"0 0 0 * * *", "2012-10-30 15:12:42", "2012-10-31 00:00:00"},
		{"0 0 0 * * *", "2012-10-31 00:00:00", "2012-11-01 00:00:00"},
		{"0 0 0 1 * *", "2012-10-30 15:12:42", "2012-11-01 00:00:00"},
		{"0 0 0 1 * *", "2012-11-01 00:00:00", "2012-12-01 00:00:00"},
		{"0 0 0 1 * *", "2010-12-31 15:12:42", "2011-01-01 00:00:00"},
		{"0 0 0 1 * *", "2011-01-01 00:00:00", "2011-02-01 00:00:00"},
		{"0 0 0 31 * *", "2011-10-30 15:12:42", "2011-10-31 00:00:00"},
		{"0 0 0 1 * *", "2011-10-30 15:12:42", "2011-11-01 00:00:00"},
		{"* * * * * 2", "2010-10-25 15:12:42", "2010-10-26 00:00:00"},
		{"* * * * * 2", "2010-10-20 15:12:42", "2010-10-26 00:00:00"},
		{"* * * * * 2", "2010-10-27 15:12:42", "2010-11-02 00:00:00"},
		{"55 5 * * * *", "2010-10-27 15:04:54", "2010-10-27 15:05:55"},
		{"55 5 * * * *", "2010-10-27 15:05:55", "2010-10-27 16:05:55"},
		{"55 * 10 * * *", "2010-10-27 09:04:54", "2010-10-27 10:00:55"},
		{"55 * 10 * * *", "2010-10-27 10:00:55", "2010-10-27 10:01:55"},
		{"* 5 10 * * *", "2010-10-27 09:04:55", "2010-10-27 10:05:00"},
		{"* 5 10 * * *", "2010-10-27 10:05:00", "2010-10-27 10:05:01"},
		{"55 * * 3 * *", "2010-10-02 10:05:54", "2010-10-03 00:00:55"},
		{"55 * * 3 * *", "2010-10-03 00:00:55", "2010-10-03 00:01:55"},
		{"* * * 3 11 *", "2010-10-02 14:42:55", "2010-11-03 00:00:00"},
		{"* * * 3 11 *", "2010-11-03 00:00:00", "2010-11-03 00:00:01"},
		{"0 0 0 29 2 *", "2007-02-10 14:42:55", "2008-02-29 00:00:00"},
		{"0 0 0 29 2 *", "2008-02-29 00:00:00", "2012-02-29 00:00:00"},
		{"0 0 7 ? * MON-FRI", "2009-09-26 00:42:55", "2009-09-28 07:00:00"},
		{"0 0 7 ? * MON-FRI", "2009-09-28 07:00:00", "2009-09-29 07:00:00"},
		{"0 30 23 30 1/3 ?", "2010-12-30 00:00:00", "2011-01-30 23:30:00"},
		{"0 30 23 30 1/3 ?", "2011-01-30 23:30:00", "2011-04-30 23:30:00"},
		{"0 30 23 30 1/3 ?", "2011-04-30 23:30:00", "2011-07-30 23:30:00"},
		{"* 6-6 * * * *", "2012-07-01 09:53:50", "2012-07-01 10:06:00"},
	}

	for i, c := range cases {
		tests = append(tests, struct {
			name       string
			expression string
			baseTime   string
			want       string
		}{
			name:       fmt.Sprintf("next_cron_%d", i),
			expression: c[0],
			baseTime:   c[1],
			want:       c[2],
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.expression, time.Local)
			if err != nil {
				t.Errorf("CronExpr.New() error = %v, expression %v", err, tt.expression)
				return
			}
			base, _ := time.Parse("2006-01-02 15:04:05", tt.baseTime)
			got, err := c.Next(base)
			if err != nil {
				t.Errorf("CronExpr.Next() error = %v, expression %v", err, tt.expression)
				return
			}
			if !reflect.DeepEqual(got.Format("2006-01-02 15:04:05"), tt.want) {
				t.Errorf("CronExpr.Next() = %v, want %v", got.Format("2006-01-02 15:04:05"), tt.want)
			}
		})
	}
}
