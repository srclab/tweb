package tweb

import "net/http"

var _ Server = &HTTPServer{}

type HandleFunc func(ctx *Context)

type Server interface {
	http.Handler
	Start(addr string) error

	// addRoute 路由注册功能
	// method 是 HTTP 方法
	// path 是路由
	// handleFunc 是你的业务逻辑
	addRoute(method string, path string, handleFunc HandleFunc)
}

type HTTPServer struct {
	router
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		router: newRouter(),
	}
}

func (s *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}
	s.serve(ctx)
}

func (s *HTTPServer) Start(addr string) error {
	return http.ListenAndServe(addr, s)
}

func (s *HTTPServer) serve(ctx *Context) {
	info, ok := s.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info.n.handler == nil {
		ctx.Resp.WriteHeader(404)
		ctx.Resp.Write([]byte("Not Found"))
		return
	}
	ctx.PathParams = info.pathParams
	info.n.handler(ctx)
}
