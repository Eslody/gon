package gon

import (
	"fmt"
	"strings"
)

type node struct {
	//匹配的路由，除叶节点外其他都为""
	path  string
	//路由中一部分
	part     string
	//子节点
	children []*node
	//是否是通配符节点
	//':'表示适配一段路径
	//'*'表示适配后面所有路径
	//这两个通配符对应的url段都不能为空
	isWild   bool
}
//读取节点内容
func (n *node) String() string {
	return fmt.Sprintf("node{path=%s, part=%s, isWild=%t}", n.path, n.part, n.isWild)
}
//递归进行路由分配
func (n *node) insert(path string, parts []string, height int) {
	//到了最后一位待匹配路由
	if len(parts) == height {
		n.path = path
		return
	}
	//这个节点匹配这个路由
	part := parts[height]
	child := n.matchChild(part)
	//尾递归也可以换成迭代
	child.insert(path, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	//递归结束条件
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.path == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)
	//最多只有一条适配的路径
	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

//深度遍历
func (n *node) travel(list *[]*node) {
	if n.path != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
	}
}
//在该节点的子节点下寻找是否有合适的匹配路径，没有就创建一个
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	//如果没有已缓存的路径
	child := &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
	n.children = append(n.children, child)
	return child
}
//找到所有合适的匹配路径
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
