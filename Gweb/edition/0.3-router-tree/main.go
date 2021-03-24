package main

import (
	"Gweb/gweb"
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
	engine.Get("/", func(c *gweb.Context) {
		c.HTML(http.StatusOK, "<h1>Hello GWeb</h1>")
	})

	engine.Get("/:name/tet", func(c *gweb.Context) {
		c.String(http.StatusOK, "hello1 %s, you are at %s\n", c.Param("name"), c.Path)
	})

	engine.Get("/hello", func(c *gweb.Context) {
		c.String(http.StatusOK, "hello0 %s, you are at %s\n", c.Query("name"), c.Path)
	})

	engine.Get("/setfile/filepath/*name", func(c *gweb.Context) {
		c.JSON(http.StatusOK, map[string]string{"filepath": c.Param("filepath")})
	})

	engine.POST("/login", func(c *gweb.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})

	log.Fatal(engine.Run(":9999"))
}
