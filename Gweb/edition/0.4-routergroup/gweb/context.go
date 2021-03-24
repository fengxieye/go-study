package gweb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Context struct {
	Writer     http.ResponseWriter
	Req        *http.Request
	Path       string
	Method     string
	StatusCode int
	Params     map[string]string
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

/*

如果err!=nil的话http.Error(c.Writer, err.Error(), 500)这里是不起作用的，因为前面已经执行了WriteHeader(code),那么返回码将不会再更改。
http.Error(c.Writer, err.Error(), 500)里面的w.WriteHeader(code)、w.Header().Set()不起作用，
而且encoder.Encode(obj)相当于调用了Write()，http.Error(c.Writer, err.Error(), 500)里面的WriteHeader、Header().Set()操作都是无效的。
gin的代码，如果encoder.Encode(obj)这里报错的话是直接panic，感觉这里如果err!=nil的话确实不好处理
*/
func (c *Context) JSON(code int, obj interface{}) error {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	//if err := encoder.Encode(obj); err != nil{
	//	http.Error(c.Writer, err.Error(), 500)
	//}
	err := encoder.Encode(obj)
	return err
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

func (c *Context) Param(key string) string {
	v, _ := c.Params[key]
	return v
}
