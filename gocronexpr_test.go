package gocronexpr

import (
	"reflect"
	"testing"
	"time"
)

func TestNext(t *testing.T) {
	str2time := func(timeStr string) time.Time {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
		return t
	}
	type args struct {
		cronexpr string
		base     time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "fields less then 6",
			args: struct {
				cronexpr string
				base     time.Time
			}{cronexpr: "0 0/2 * * ?", base: time.Now()},
			want:    time.Now(),
			wantErr: true,
		},
		{
			name: "run at every hour",
			args: struct {
				cronexpr string
				base     time.Time
			}{cronexpr: "0 0 0/1 * * ?", base: str2time("2019-01-01 00:00:01")},
			want:    str2time("2019-01-01 01:00:00"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Next(tt.args.cronexpr, tt.args.base)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Next() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}
