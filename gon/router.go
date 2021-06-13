package gon

import (
	"net/http"
	"strings"
)

type router struct {
	//请求方式作为根节点如["GET"]
	roots    map[string]*node
	//存储请求方式-路由所对应的HandlerFunc
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots: make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}
//路径解析，只保留有效部分，即：如果是*，只会保留*后/前的内容，/后的内容都会舍弃
func parsePath(path string) []string {
	vs := strings.Split(path, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, path string, handler HandlerFunc) {
	parts := parsePath(path)
	key := method + "-" + path

	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	//比如：GET方法增加一条路由
	r.roots[method].insert(path, parts, 0)

	r.handlers[key] = handler
}
//获得适配的前缀树节点和映射规则
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePath(path)	//{"hello","gonzo"}
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePath(n.path)	//{"hello",":name"}
		for index, part := range parts {
			if part[0] == ':' {
				//name: gonzo
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {	//part不为一个单独的"*"
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}
//获得一个方法下所有路由节点
func (r *router) getRoutes(method string) []*node {
	root, ok := r.roots[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)

	if n!= nil {
		c.Params = params
		key := c.Method + "-" + n.path
		//添加处理函数
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
	//顺次执行中间件——路由handler
	c.Next()
}