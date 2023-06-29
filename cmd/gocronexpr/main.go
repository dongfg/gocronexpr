package main

import (
	"flag"
	"fmt"
	"github.com/dongfg/gocronexpr"
	"os"
	"strconv"
	"time"
)

type color string

const (
	colorRed    color = "\u001b[31m"
	colorGreen        = "\u001b[32m"
	colorYellow       = "\u001b[33m"
	colorReset        = "\u001b[0m"
)

func init() {
	flag.Usage = func() {
		fmt.Printf("NAME:\n  %s\n", "gocronexpr - Display the time of the next N runs base on cron expression")
		fmt.Printf("USAGE:\n  %s\n", "gocronexpr <cron> [N]")
		fmt.Printf("OPTIONS:\n")
		fmt.Printf("  %-8s%s\n", "<cron>", "6 fields cron expression")
		fmt.Printf("  %-8s%s\n", "[N]", "next number of runs, default 5")
		os.Exit(0)
	}
}

func colorize(color color, message string) {
	fmt.Printf("%s%s%s\n", string(color), message, colorReset)
}

func parse() (cron string, times int, err error) {
	cron = os.Args[1]
	if len(os.Args) == 2 {
		times = 5
	} else {
		_t, err := strconv.Atoi(os.Args[2])
		if err != nil {
			return cron, times, err
		}
		times = _t
	}
	return cron, times, err
}

func main() {
	if len(os.Args) == 2 && len(os.Args) == 3 {
		flag.Usage()
	}
	cron, times, err := parse()
	if err != nil {
		colorize(colorRed, fmt.Sprintf("Error: %+v", err))
		flag.Usage()
	}

	cronExpr, err := gocronexpr.New(cron, time.Local)
	if err != nil {
		colorize(colorRed, fmt.Sprintf("Error: %+v", err))
		return
	}

	base := time.Now()
	for i := 0; i < times; i++ {
		nextTime, err := cronExpr.Next(base)
		if err != nil {
			colorize(colorRed, err.Error())
			return
		}
		base = nextTime
		colorize(colorGreen, fmt.Sprintf("%*d: %s%s", len(strconv.Itoa(times)), i+1,
			colorYellow, nextTime.Format("2006-01-02 15:04:05")))
	}
}
