# prue
http web mvc-framework written in Golang.

### Getting Prue
```shell
import "github.com/geekip/prue"
```
```shell
go get -u github.com/geekip/prue"
```

Cookie
session
Redis
Static Server
404
orm
csrf
validator


https://cloud.tencent.com/developer/article/1422442
https://github.com/mydevc/go-gin-mvc/tree/master

https://github.com/wangweimei/gomvc
https://github.com/z924931408/go-admin

https://github.com/xujiajun/gorouter


使用golan编写http路由，需求如下
1、使用trie算法(数据结构压缩)
2、支持常规路由、正则路由、子路由
3、原生go实现，不需要第三方库
4、能自定义NotFoundHandler 404 和InternalError 500
5、支持中间件Middleware 
6、支持文件服务 Static Files Server
7、参考：https://github.com/xujiajun/gorouter https://github.com/gorilla/mux
7、路由使用demo如下

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello world!"))
}

func FilesHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello world!"))
}

func main(){
	// 实例化
	router := NewRouter()
	// 普通路由
	router.Path("/").Handler(HomeHandler)
	// 支持传入Methods
	router.Path("/about").Handler(AboutHandler).Methods("GET","POST")
	// 子路由
	subRouter:=router.Group("/news").Subrouter()
	subRouter.Path("/list").Handler(NewListHandler).Methods("GET")
	// 正则路由
	subRouter.Path("/edit/{id}").Handler(NewEditHandler).Methods("POST")
	subRouter.Path("/view/{id:[0-9]+}").Handler(NewViewHandler)
	// 路由命名
	router.Path("/test").Handler(HomeHandler).Methods("GET","POST").Name("Test")
	// 中间件
	router.Use(Middleware,Middleware2)

	// 文件服务
	filePath:= "/files/"
	fileDir := http.Dir("./files/")
	fileServer := http.FileServer(fileDir)
	fileHander := http.StripPrefix(filePath, fileServer)
	router.Group(filePath).Handler(fileHander)

	log.Fatal(http.ListenAndServe("0.0.0.0:3000", router))
}

