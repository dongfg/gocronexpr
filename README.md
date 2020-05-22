[![GoDoc](http://godoc.org/github.com/dongfg/gocronexpr?status.png)](http://godoc.org/github.com/dongfg/gocronexpr) 
[![Build Status](https://travis-ci.org/dongfg/gocronexpr.svg?branch=master)](https://travis-ci.org/dongfg/gocronexpr)
Cron expression parsing in go
=================================
### Introduction
Supports cron expressions with `seconds` field. A copy of [CronSequenceGenerator](https://github.com/spring-projects/spring-framework/blob/fd48bf1dbe9d7d619cd9e076d5f5bc60659c25a3/spring-context/src/main/java/org/springframework/scheduling/support/CronSequenceGenerator.java#L84) from Spring Framework.

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

func main()  {
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
