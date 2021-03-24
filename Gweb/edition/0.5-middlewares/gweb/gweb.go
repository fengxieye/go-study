package gweb

import (
	"net/http"
	"strings"
)

type HandleFunc func(ctx *Context)

type Engine struct {
	router *router
	*RouterGroup
	groups []*RouterGroup
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

//移入 routergroup 中
//func (engine *Engine) addRouter(method string, pattern string, handler HandleFunc)  {
//	engine.router.addRouter(method, pattern, handler)
//}
//
//func (engine *Engine) Get (pattern string, handler HandleFunc){
//	engine.addRouter("GET", pattern, handler)
//}
//
//func (engine *Engine) POST (pattern string, handler HandleFunc){
//	engine.addRouter("POST", pattern, handler)
//}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandleFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	engine.router.Handle(c)
}
