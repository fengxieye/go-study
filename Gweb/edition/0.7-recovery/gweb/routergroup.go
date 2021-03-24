package gweb

import (
	"net/http"
	"path"
)

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

func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandleFunc {
	absolutePath := path.Join(g.prefix, relativePath)
	//会去掉absolutePath，就是路由部分，用真实的路径代替
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filePath")
		//log.Printf("file:%s", file)
		if _, err := fs.Open(file); err != nil {
			//log.Printf("not find file %s", file)
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

func (g *RouterGroup) Static(relativePath string, root string) {
	handler := g.createStaticHandler(relativePath, http.Dir(root))
	//取文件对应到get方法，放入路由
	urlPattern := path.Join(relativePath, "/*filePath")
	g.Get(urlPattern, handler)
}
