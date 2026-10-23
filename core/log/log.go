package log

import (
	"fmt"
	"runtime"
)

func Println(args ...interface{}) {
	c, f, l, ok := runtime.Caller(1)
	if !ok {
		fmt.Println(args...)
	}
	fn := runtime.FuncForPC(c)
	fmt.Println(fmt.Sprintf("%s():%d in %s", fn.Name(), l, f))
	fmt.Println(args...)
}
