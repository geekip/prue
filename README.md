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
https://github.com/mydevc/go-gin-mvc/blob/master/middleware/csrf/csrf.go

validator

前缀树算法实现路由匹配原理解析
https://github.com/ischaojie/gaga

gin框架中间件详解
https://www.jb51.net/article/281431.htm

基于gin框架的web开发脚手架
https://zhuanlan.zhihu.com/p/645008522

基于框架gin+xorm搭建的MVC项目
https://cloud.tencent.com/developer/article/1422442
https://github.com/mydevc/go-gin-mvc/tree/master

简单的高性能 Golang MVC 框架
https://github.com/wangweimei/gomvc

基于Gin+gorm框架搭建的MVC
https://github.com/z924931408/go-admin

https://github.com/xujiajun/gorouter

Go语言相关的话题分享
https://github.com/talkgo/night

https://github.com/julienschmidt/httprouter
https://github.com/go-chi/chi
https://github.com/gorilla/mux
https://github.com/bmizerany/pat
https://github.com/matryer/way
https://github.com/alexedwards/flow
https://github.com/celrenheit/lion
https://github.com/gmosx/servemux
https://github.com/claygod/Bxog
https://github.com/donutloop/mux/
https://github.com/gernest/alien
https://github.com/go-ozzo/ozzo-routing
https://www.alexedwards.net/blog/which-go-router-should-i-use
路由集合
https://mp.weixin.qq.com/s/RnRkS8stazKlMN6aFAdtwA


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
