package v3

import (
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	s := NewHTTPServer()

	s.addRoute(http.MethodGet, "/user", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, user"))
	})

	s.Start(":8081")
}
