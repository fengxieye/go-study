package gweb

type RouterGroup struct {
	prefix      string
	middlewares []HandleFunc
	parent      *RouterGroup
	engine      *Engine
}

func (g *RouterGroup) Group(prefix string) *RouterGroup {
	engine := g.engine
	newGroup := &RouterGroup{
		prefix: g.prefix + prefix,
		parent: g,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (g *RouterGroup) addRouter(method string, comp string, handler HandleFunc) {
	pattern := g.prefix + comp
	//log.Printf("Router %4s - %s", method, pattern)
	g.engine.router.addRouter(method, pattern, handler)
}

func (g *RouterGroup) Get(pattern string, handler HandleFunc) {
	g.addRouter("GET", pattern, handler)
}

func (g *RouterGroup) POST(pattern string, handler HandleFunc) {
	g.addRouter("POST", pattern, handler)
}

func (g *RouterGroup) Use(middlewares ...HandleFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}
