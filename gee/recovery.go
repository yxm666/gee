package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// recovery 也作为一种 中间件
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next()
	}
}

// 获取触发 panic 的堆栈信息
func trace(message string) string {
	var pcs [32]uintptr
	runtime.Callers(3, pcs[:])
	/*	调用了Callers 用来返回调用栈的程序计数器,
		第 0 个 Caller 是 Callers 本身，第 1 个是上一层 trace，第 2 个是再上一层的 defer func。
		因此，为了日志简洁一点，我们跳过了前 3 个 Caller。*/
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		//返回一个表示调用栈标识符pc对应的调用栈的*Func 获取对应的函数
		fn := runtime.FuncForPC(pc)
		// 获取调用该函数的文件名和行号，打印在日志中
		//FileLine 返回该调用栈所调用的函数的源代码文件名和行号。如果pc不是f内的调用栈标识符，结果是不精确的
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
