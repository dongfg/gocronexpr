Cron expression parsing in go
=================================
[![GoDoc](http://godoc.org/github.com/dongfg/gocronexpr?status.png)](http://godoc.org/github.com/dongfg/gocronexpr)
[![Test Status](https://github.com/dongfg/gocronexpr/workflows/Go%20Test/badge.svg)](https://github.com/dongfg/gocronexpr/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/dongfg/gocronexpr)](https://goreportcard.com/report/github.com/dongfg/gocronexpr)

### Introduction

Supports cron expressions with `seconds` field. A copy
of [CronSequenceGenerator](https://github.com/spring-projects/spring-framework/blob/fd48bf1dbe9d7d619cd9e076d5f5bc60659c25a3/spring-context/src/main/java/org/springframework/scheduling/support/CronSequenceGenerator.java#L84)
from Spring Framework.

### Installation

The import path for the package is `github.com/dongfg/gocronexpr`.

To install it, run:

```shell
 go get github.com/dongfg/gocronexpr
 ```

### License

The package is licensed under the Apache License 2.0. Please see the LICENSE file for details.

### Example

```go
package main

import (
	"fmt"
	"github.com/dongfg/gocronexpr"
	"time"
)

func main() {
	cronExpr, err := gocronexpr.New("* * * * * *", time.Local)
	if err != nil {
		fmt.Println(err)
		return
	}

	nextTime, err := cronExpr.Next(time.Now())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(nextTime)
}
```

This example will generate the following output:

```text
2019-05-22 17:13:22 +0800 CST
```

### Install as a command

Download from [release](https://github.com/dongfg/gocronexpr/releases) page or install with go command:

```shell
GO111MODULE="off" go get github.com/dongfg/gocronexpr/cmd/gocronexpr
```

Usage:
```shell
dongfg at MacBook-Pro.local in [~]
10:42:08 $ gocronexpr
NAME:
  gocronexpr - Display the time of the next N runs base on cron expression
USAGE:
  gocronexpr <cron> [N]
OPTIONS:
  <cron>  6 fields cron expression
  [N]     next number of runs, default 5

dongfg at MacBook-Pro.local in [~]
10:42:10 $ gocronexpr "5 */5 * * * *"
1: 2021-07-14 10:45:05
2: 2021-07-14 10:50:05
3: 2021-07-14 10:55:05
4: 2021-07-14 11:00:05
5: 2021-07-14 11:05:05

dongfg at MacBook-Pro.local in [~]
10:42:15 $ gocronexpr "5 */5 * * * *" 3
1: 2021-07-14 10:45:05
2: 2021-07-14 10:50:05
3: 2021-07-14 10:55:05
```