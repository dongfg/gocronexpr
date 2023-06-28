package main

import (
	"flag"
	"fmt"
	"github.com/dongfg/gocronexpr"
	"os"
	"strconv"
	"time"
)

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		_, _ = fmt.Fprintf(w, "NAME:\n  %s\n", "gocronexpr - Display the time of the next N runs base on cron expression")
		_, _ = fmt.Fprintf(w, "USAGE:\n  %s\n", "gocronexpr <cron> [N]")
	}

	if len(os.Args) < 1 || len(os.Args) > 3 {
		flag.Usage()
	}

	cron := os.Args[1]
	times, err := strconv.Atoi(os.Args[2])
	if err != nil {
		flag.Usage()
	}
	
	cronExpr, err := gocronexpr.New(cron, time.Local)
	if err != nil {
		fmt.Println(err)
		return
	}

	base := time.Now()
	for i := 0; i < times; i++ {
		nextTime, err := cronExpr.Next(base)
		if err != nil {
			fmt.Println(err)
			return
		}
		base = nextTime
		fmt.Println(nextTime)
	}
}
