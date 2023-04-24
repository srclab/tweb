package v1

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
							},
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
