//日志中间件
package gon

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {

		t := time.Now()
		//处理请求
		c.Next()

		log.Printf("[%d] %s in %v", c.writermem.status, c.Request.RequestURI, time.Since(t))
	}
}

