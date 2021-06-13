//错误处理中间件
package gon

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				//打印错误日志
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}

//获取panic堆栈信息
func trace(message string) string {
	var pcs [32]uintptr
	//Callers用于返回调用栈的PC，第0个是callers本身，第一个是trace，第二个是调用者defer func
	n := runtime.Callers(3, pcs[:])
	log.Println(pcs[0])
	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	//所有调用函数
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
