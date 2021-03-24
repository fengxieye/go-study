package gweb

import (
	"log"
	"net/http"
)

type router struct {
	handles map[string]HandleFunc
}

func newRouter() *router {
	return &router{handles: map[string]HandleFunc{}}
}

func (r *router) AddRouter(method string, pattern string, handler HandleFunc) {
	log.Printf("Router %4s - %s", method, pattern)
	key := method + "-" + pattern
	r.handles[key] = handler
}

func (r *router) Handle(c *Context) {
	key := c.Method + "-" + c.Path
	if handler, ok := r.handles[key]; ok {
		handler(c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
