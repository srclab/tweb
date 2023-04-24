package v1

import (
	"strings"
)

// router 代表路由树（森林）
type router struct {
	// map[http.Method]路由树根节点
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {

	root, ok := r.trees[method]
	if !ok {
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	//根节点特殊处理一下
	if path == "/" {
		root.handler = handleFunc
		return
	}

	// /user/home 被切割成三段
	segs := strings.Split(path[1:], "/")
	for _, seg := range segs {
		// 递归下去，找准位置
		// 如果中途有节点不存在，需要创建出来
		child := root.childGetOrCreate(seg)
		root = child
	}
	root.handler = handleFunc
}

func (r *router) findRoute(method string, path string) (*node, bool) {
	return nil, false
}

type node struct {
	path string
	// map[子 path]子节点
	children map[string]*node
	// handler 命中路由之后执行的逻辑
	handler HandleFunc
}

func (n *node) childGetOrCreate(seg string) *node {
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[seg]
	if !ok {
		child = &node{
			path: seg,
		}
		n.children[seg] = child
	}
	return child
}
