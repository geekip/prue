package prue

import (
	"net/http"
)

type Handler func(ctx *Context)

type Router struct {
	prefix        string
	trie          *trie
	notFound      Handler
	internalError func(ctx *Context, err interface{})
	middlewares   []Handler
}

func NewRouter() *Router {
	return &Router{
		trie:          newTrie(),
		notFound:      defaultNotFound,
		internalError: defaultInternalError,
	}
}

func (r *Router) Use(middleware ...Handler) *Router {
	r.middlewares = append(r.middlewares, middleware...)
	return r
}

func (r *Router) PathPrefix(pattern string) *Router {
	return &Router{
		prefix:      pattern,
		trie:        r.trie,
		middlewares: r.middlewares,
	}
}

func (r *Router) Handle(method, pattern string, handler Handler) *Router {
	pattern = r.prefix + "/" + pattern
	r.trie.add(method, pattern, handler, r.middlewares)
	return r
}

func defaultNotFound(ctx *Context) {
	http.Error(ctx.Response, "404 page not founds", http.StatusNotFound)
}

func defaultInternalError(ctx *Context, err interface{}) {
	http.Error(ctx.Response, "500 internal server error", http.StatusInternalServerError)
}

func (r *Router) NotFound(handler Handler) {
	r.notFound = handler
}

func (r *Router) InternalError(handler func(ctx *Context, err interface{})) {
	r.internalError = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := newContext(w, req)

	defer func() {
		if err := recover(); err != nil {
			r.internalError(ctx, err)
		}
		releaseContext(ctx)
	}()

	node := r.trie.find(req.Method, ctx.Path)
	if node == nil {
		r.notFound(ctx)
		return
	}
	ctx.Params = node.params
	ctx.handlers = node.handlers
	ctx.Next()
}

func (r *Router) ALL(pattern string, handle Handler) *Router {
	return r.Handle(wildcardPrefix, pattern, handle)
}

func (r *Router) GET(pattern string, handler Handler) *Router {
	return r.Handle(http.MethodGet, pattern, handler)
}

func (r *Router) HEAD(pattern string, handler Handler) *Router {
	return r.Handle(http.MethodHead, pattern, handler)
}

func (r *Router) POST(pattern string, handler Handler) *Router {
	return r.Handle(http.MethodPost, pattern, handler)
}

func (r *Router) PUT(pattern string, handler Handler) *Router {
	return r.Handle(http.MethodPut, pattern, handler)
}

func (r *Router) PATCH(pattern string, handler Handler) *Router {
	return r.Handle(http.MethodPatch, pattern, handler)
}

func (r *Router) OPTIONS(pattern string, handler Handler) *Router {
	return r.Handle(http.MethodOptions, pattern, handler)
}

func (r *Router) DELETE(pattern string, handle Handler) *Router {
	return r.Handle(http.MethodDelete, pattern, handle)
}

func (r *Router) TRACE(pattern string, handler Handler) *Router {
	return r.Handle(http.MethodTrace, pattern, handler)
}
