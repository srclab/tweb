package tweb

import (
	"fmt"
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

// addRoute 注册路由
// method 是 HTTP 方法
// path 必须以 / 开始并且结尾不能有 /，中间也不允许有连续的 /
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	if path == "" {
		panic("web: 路由是空字符串")
	}
	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}
	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}

	root, ok := r.trees[method]
	if !ok {
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	//根节点特殊处理一下
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突[/]")
		}
		root.handler = handleFunc
		return
	}

	// /user/home 被切割成三段
	segs := strings.Split(path[1:], "/")
	for _, seg := range segs {
		if seg == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		// 递归下去，找准位置
		// 如果中途有节点不存在，需要创建出来
		child := root.childGetOrCreate(seg)
		root = child
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	root.handler = handleFunc
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}

// findRoute 查找对应的节点
// 注意，返回的 node 内部 HandleFunc 不为 nil 才算是注册了路由
func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return &matchInfo{n: root}, true
	}

	segs := strings.Split(strings.Trim(path, "/"), "/")
	var pathParams map[string]string
	for _, seg := range segs {
		var matchParam bool
		root, matchParam, ok = root.childGet(seg)
		if !ok {
			return nil, false
		}
		if matchParam {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			pathParams[root.path[1:]] = seg
		}
	}
	return &matchInfo{
		n:          root,
		pathParams: pathParams,
	}, true
}

type node struct {
	path string
	// map[子 path]子节点
	children map[string]*node
	// 通配符匹配
	starChild *node
	// 参数路径
	paramChild *node
	// handler 命中路由之后执行的逻辑
	handler HandleFunc
}

func (n *node) childGetOrCreate(seg string) *node {
	if seg[0] == ':' {
		if n.starChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", seg))
		}
		n.paramChild = &node{path: seg}
		return n.paramChild
	}

	if seg == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", seg))
		}
		n.starChild = &node{path: seg}
		return n.starChild
	}

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

// child 返回子节点
// 第一个返回值 *node 是命中的节点
// 第二个返回值 bool 代表是否是命中参数路由
// 第三个返回值 bool 代表是否命中
func (n *node) childGet(seg string) (*node, bool, bool) {
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	child, ok := n.children[seg]
	if !ok {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	return child, false, ok
}
