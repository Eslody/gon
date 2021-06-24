package gon

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	//封装状态码和其他一些数据
	writermem responseWriter
	Writer	 ResponseWriter
	Request    *http.Request
	//url及方法
	Path   string
	Method string
	//动态路由映射
	Params map[string]string
	//处理函数
	handlers []HandlerFunc
	index    int
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	c := &Context{
		writermem: responseWriter{},
		Request:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index: -1,
	}
	c.writermem.reset(w)
	c.Writer = &c.writermem

	return c
}
//顺次执行所有中间件，类似栈，index作为全局变量记录
func (c *Context) Next()  {
	c.index++
	for ; c.index < len(c.handlers); c.index++ {
		c.handlers[c.index](c)
	}
}
//所有handlers直接结束
func (c *Context) Abort() {
	c.index = len(c.handlers)
}

//获取post内容中key对应的value
func (c *Context) PostForm(key string) string {
	return c.Request.FormValue(key)
}
//获取get方法附加的参数，如：/hello?name=gonzo
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) WriteStatus(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}
//构造response体各种类型的方法
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.WriteStatus(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.WriteStatus(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.WriteStatus(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.WriteStatus(code)
	c.Writer.Write([]byte(html))
}

