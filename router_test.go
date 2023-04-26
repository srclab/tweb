package tweb

import (
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_router_addRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		// 通配符测试用例
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		// 参数路径
		{
			method: http.MethodGet,
			path:   "/order/detail/:id",
		},
	}

	var mockHandler HandleFunc = func(ctx *Context) {}
	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path: "/",
				children: map[string]*node{
					"user": {
						path: "user",
						children: map[string]*node{
							"home": {
								path:    "home",
								handler: mockHandler,
							},
						},
						handler: mockHandler,
					},
					"order": {
						path: "order",
						children: map[string]*node{
							"detail": {
								path:    "detail",
								handler: mockHandler,
								paramChild: &node{
									path:    ":id",
									handler: mockHandler,
								},
							},
						},
						starChild: &node{
							path:    "*",
							handler: mockHandler,
						},
						handler: nil,
					},
				},
				handler: mockHandler,
			},
			http.MethodPost: {
				path: "/",
				children: map[string]*node{
					"order": {
						path: "order",
						children: map[string]*node{
							"create": {
								path:    "create",
								handler: mockHandler,
							},
						},
						handler: nil,
					},
					"login": {
						path:    "login",
						handler: mockHandler,
					},
				}},
		},
	}
	err := wantRouter.equal(r)
	assert.True(t, err == nil, err)
	// // 因为有 function 字段，所以 Equal 无法比较。
	// assert.Equal(t, wantRouter, r)

	// 非法用例
	r = newRouter()

	// 空字符串
	assert.PanicsWithValue(t, "web: 路由是空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})

	// 前导没有 /
	assert.PanicsWithValue(t, "web: 路由必须以 / 开头", func() {
		r.addRoute(http.MethodGet, "a/b/c", mockHandler)
	})

	// 后缀有 /
	assert.PanicsWithValue(t, "web: 路由不能以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	})

	// 根节点重复注册
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/]", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})
	// 普通节点重复注册
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/a/b/c]", func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})

	// 多个 /
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [/a//b]", func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [//a/b]", func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	})

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/*", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	}, "web: 不允许同时注册路径参数和通配符匹配，已有通配符匹配")

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
	}, "web: 不允许同时注册路径参数和通配符匹配，已有路径参数")
}

func (want router) equal(get router) error {
	for method, wantRoot := range want.trees {
		getRoot := get.trees[method]
		if err := wantRoot.equal(getRoot, ""); err != nil {
			return err
		}
	}
	return nil
}

func (want *node) equal(get *node, id string) error {
	var path string
	if want != nil {
		path = want.path
	}
	id = filepath.Join(id, path)
	if want == nil || get == nil {
		return fmt.Errorf("node(id=%q) is nil", id)
	}

	if want.starChild != nil {
		if err := want.starChild.equal(get.starChild, id); err != nil {
			return err
		}
	}

	if want.paramChild != nil {
		if err := want.paramChild.equal(get.paramChild, id); err != nil {
			return err
		}
	}

	if want.path != get.path {
		return fmt.Errorf("node(id=%q) path want: %q, get: %q", id, want.path, get.path)
	}

	if len(want.children) != len(get.children) {
		return fmt.Errorf("node(id=%q) len(children) want: %d, get: %d", id, len(want.children), len(get.children))
	}

	wantHandler := reflect.ValueOf(want.handler)
	getHandler := reflect.ValueOf(get.handler)
	if wantHandler != getHandler {
		return fmt.Errorf("node(id=%q) handler want: %+v, get: %+v", id, wantHandler, getHandler)
	}

	for childPath, wantChildNode := range want.children {
		getChildNode := get.children[childPath]
		if err := wantChildNode.equal(getChildNode, id); err != nil {
			return err
		}
	}

	return nil
}

func Test_router_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
		{
			method: http.MethodPost,
			path:   "/order/create/:id",
		},
	}

	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name   string
		method string
		path   string
		found  bool
		info   *matchInfo
	}{
		{
			name:   "method not found",
			method: http.MethodHead,
		},
		{
			name:   "path not found",
			method: http.MethodGet,
			path:   "/abc",
		},
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
			found:  true,
			info: &matchInfo{
				n: &node{
					path: "/",
					children: map[string]*node{
						"user": {
							path:    "user",
							handler: mockHandler,
						},
					},
					handler: mockHandler,
				},
			},
		},
		{
			name:   "user",
			method: http.MethodGet,
			path:   "/user",
			found:  true,
			info: &matchInfo{
				n: &node{
					path:    "user",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "no handler",
			method: http.MethodPost,
			path:   "/order",
			found:  true,
			info: &matchInfo{
				n: &node{
					path: "order",
					children: map[string]*node{
						"create": {
							path:    "create",
							handler: mockHandler,
							paramChild: &node{
								path:    ":id",
								handler: mockHandler,
							},
						},
					},
				},
			},
		},
		{
			// 完全命中
			name:   "two layer",
			method: http.MethodPost,
			path:   "/order/create",
			found:  true,
			info: &matchInfo{
				n: &node{
					path:    "create",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "order star",
			method: http.MethodPost,
			path:   "/order/abc",
			found:  true,
			info: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// id 路径参数匹配
			name:   "create id",
			method: http.MethodPost,
			path:   "/order/create/123",
			found:  true,
			info: &matchInfo{
				n: &node{
					path:    ":id",
					handler: mockHandler,
				},
				pathParams: map[string]string{
					"id": "123",
				},
			},
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getInfo, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			assert.Equal(t, tc.info.pathParams, getInfo.pathParams)
			err := tc.info.n.equal(getInfo.n, "")
			assert.True(t, err == nil, err)
		})
	}
}
