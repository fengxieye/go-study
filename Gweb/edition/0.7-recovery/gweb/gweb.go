package gweb

import (
	"html/template"
	"net/http"
	"strings"
)

type HandleFunc func(ctx *Context)

type Engine struct {
	router *router
	*RouterGroup
	groups []*RouterGroup
	//模板化
	htmlTemplates *template.Template
	funcMap       template.FuncMap
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
	c.engine = engine
	engine.router.Handle(c)
}

//模板函数
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

//读取所有模板保存
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}
