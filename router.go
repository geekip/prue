package prue

import (
	"net/http"

	"github.com/gorilla/mux"
)

// 注册中间件
func applyMiddlewares(ctx *Context, handler Handler, middlewares Middlewares) {
	if len(middlewares) == 0 {
		handler(ctx)
		return
	}

	var next Handler
	next = func(ctx *Context) {
		if len(middlewares) == 0 {
			handler(ctx)
			return
		}
		middleware := middlewares[0]
		middlewares = middlewares[1:]
		middleware(ctx, next)
	}

	next(ctx)
}

// 创建上下文实例
func makeHandler(handlerFunc Handler, middlewares Middlewares) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := &Context{
			Data:     make(map[string]interface{}),
			Request:  request,
			Response: response,
		}
		applyMiddlewares(ctx, handlerFunc, middlewares)
	}
}

// 注册路由
func mountRoutes(router *mux.Router, routes Routes) {
	for _, route := range routes {
		handler := makeHandler(route.Handler, route.Middlewares)
		router.Methods(route.Methods...).Path(route.Path).Handler(handler).Name(route.Name)
		if len(route.SubRoutes) > 0 {
			subRouter := router.PathPrefix(route.Path).Subrouter()
			mountRoutes(subRouter, route.SubRoutes)
		}
	}
}
