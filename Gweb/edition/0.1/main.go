package main

import (
	"Gweb/gweb"
	"fmt"
	"log"
	"net/http"
)

//type Engine struct {
//
//}
//
//func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request)  {
//	switch req.URL.Path {
//	case "/":
//		fmt.Fprintf(w, "URL.Path=%q\n", req.URL.Path)
//	case "/hello":
//		for k,v := range req.Header {
//			fmt.Fprintf(w, "header[%q] = %q\n", k, v)
//		}
//	default:
//		fmt.Fprintf(w, "404 NOT FOUND:%s\n", req.URL)
//	}
//}

//第一步封装启动函数，实现静态路由表，注册方法
func main() {
	//engine := new(Engine)
	//	//log.Fatal(http.ListenAndServe(":9999", engine))

	engine := gweb.New()
	engine.Get("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "URL.Path=%q\n", req.URL.Path)
	})

	engine.Get("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			fmt.Fprintf(w, "header[%q] = %q\n", k, v)
		}
	})

	log.Fatal(engine.Run(":9999"))
}
