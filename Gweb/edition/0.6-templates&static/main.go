package main

import (
	"Gweb/gweb"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Logger() gweb.HandleFunc {
	return func(ctx *gweb.Context) {
		t := time.Now()
		ctx.Next()
		log.Printf("[%d] %s in %v", ctx.StatusCode, ctx.Req.RequestURI, time.Since(t).Microseconds())
	}
}

func FormatAsData(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

//第一步封装启动函数，实现静态路由表，注册方法
func main() {
	engine := gweb.New()
	engine.Use(Logger())

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		strings.Replace(dir, "\\", "/", -1)
		dir = dir + "/static"
	}
	//访问 http://localhost:9999/assets/test.js
	engine.Static("/assets", dir)

	engine.SetFuncMap(template.FuncMap{
		"FormatAsData": FormatAsData,
	})

	//http://localhost:9999/assets/test.js
	engine.LoadHTMLGlob("templates/*")
	engine.Get("/", func(c *gweb.Context) {
		c.HTML(http.StatusOK, "test.html", nil)
	})

	log.Fatal(engine.Run(":9999"))
}
