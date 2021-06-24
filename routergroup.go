package gon

import (
	"log"
	"net/http"
	"path"
)

type RouterGroup struct {
	//分组对应前缀
	prefix      string
	//支持的中间件
	middlewares []HandlerFunc
	//parent group
	parent      *RouterGroup
	engine      *Engine
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

//分组添加路由
func (group *RouterGroup) addRoute(method string, path string, handler HandlerFunc) {
	path = group.prefix + path
	log.Printf("Route %4s - %s", method, path)
	group.engine.router.addRoute(method, path, handler)
}

func (group *RouterGroup) GET(path string, handler HandlerFunc) {
	group.addRoute("GET", path, handler)
}

func (group *RouterGroup) POST(path string, handler HandlerFunc) {
	group.addRoute("POST", path, handler)
}
//在分组下使用某中间件
func (group * RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc{
	absolutePath := path.Join(group.prefix, relativePath)	//路由绝对路径
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))	//这个文件服务器存放静态文件
	return func(c *Context) {
		file := c.Params["filepath"]
		//验权
		if _, err := fs.Open(file); err != nil {
			c.WriteStatus(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}


//将磁盘上某文件夹root映射到路由relativePath
//如 r.Static("/assets", "/usr/web/static")
//或相对路径 r.Static("/assets", "./static")
//r.Run(":9999")
//用户访问localhost:9999/assets/js/gon.js，最终返回/usr/web/static/js/gon.js。
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))	//转成文件类型实现了open方法
	urlPattern := path.Join(relativePath, "/*filepath")
	//注册GET方法
	group.GET(urlPattern, handler)
}