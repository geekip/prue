package prue

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// 控制器
type Handler func(ctx *Context)

// 中间件
type Middlewares []func(Handler) Handler

// 路由
type Routes []struct {
	Name        string
	Path        string
	Methods     []string
	Handler     Handler
	Middlewares Middlewares
	SubRoutes   Routes
}

// 注册中间件
func applyMiddlewares(handler Handler, middlewares Middlewares) Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

// 注册路由
func mountRoute(router *mux.Router, routes Routes) {
	for _, route := range routes {

		handler := applyMiddlewares(route.Handler, route.Middlewares)
		router.Methods(route.Methods...).Path(route.Path).Handler(makeHandler(handler)).Name(route.Name)

		if len(route.SubRoutes) > 0 {
			subRouter := router.PathPrefix(route.Path).Subrouter()
			mountRoute(subRouter, route.SubRoutes)
		}
	}
}

// 开启服务
func (routes *Routes) Run(port int) error {
	if port == 0 {
		port = 80
	}
	var portString string = ":" + strconv.Itoa(port)
	router := mux.NewRouter()
	mountRoute(router, *routes)
	return http.ListenAndServe(portString, router)
}
