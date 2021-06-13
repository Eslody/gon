package gon

import (
	"net/http"
	"strings"
)
//静态路由处理函数
type HandlerFunc func(c *Context)

//抽象engine作为最顶层分组，继承分组所有能力
type Engine struct {
	router *router
	//结构体嵌套，这里也可以用指针
	RouterGroup
	//储存所有分组
	groups []*RouterGroup
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = RouterGroup{engine: engine}
	//指向自身的RouterGroup
	engine.groups = []*RouterGroup{&engine.RouterGroup}
	return engine

}
//初始化中间件一系列操作
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//通过URL前缀判断该请求适用于哪些中间件
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	//处理函数添加中间件
	c := newContext(w, req)
	c.handlers = middlewares
	engine.router.handle(c)
}
