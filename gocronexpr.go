package gocronexpr

import "C"
import (
	"errors"
	"time"
	"unsafe"
)

// #cgo CFLAGS: -I./ccronexpr
// #cgo LDFLAGS: -L./ccronexpr -lccronexpr
// #include <stdlib.h>
// #include <time.h>
// #include "ccronexpr.h"
import "C"

// Next time calculated based on the given time.
func Next(expression string, base time.Time) (time.Time, error) {
	var expr = C.CString(expression)
	defer C.free(unsafe.Pointer(expr))

	var cronExpr C.cron_expr
	var err = C.CString("")
	defer C.free(unsafe.Pointer(err))
	C.cron_parse_expr(expr, &cronExpr, &err)
	if err != nil {
		return base, errors.New(C.GoString(err))
	}

	cTime := C.cron_next(&cronExpr, (C.time_t)(base.Unix()))

	return time.Unix(int64(cTime), 0), nil
}
