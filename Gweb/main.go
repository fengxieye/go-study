package main

import (
	"Gweb/gweb"
	"fmt"
	"log"
	"net/http"
	"time"
)

func FormatAsData(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

//第一步封装启动函数，实现静态路由表，注册方法
func main() {
	engine := gweb.Default()

	engine.Get("/testpanic", func(c *gweb.Context) {
		names := []string{"gwebtest"}
		c.String(http.StatusOK, names[100])
	})

	log.Fatal(engine.Run(":9999"))
}
