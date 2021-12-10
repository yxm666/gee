package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//H 在golang中 interface{} 代表obj
type H map[string]interface{}

// Context 除了封装response request 还包括path method statusCode等
type Context struct {
	Writer http.ResponseWriter
	Req    *http.Request
	Path   string
	Method string
	// 将解析后的参数存储到Params中  eg:/login/:name  params name: XXX
	Params     map[string]string
	StatusCode int
	// 作为中间件 接受到请求后，应查找所有应用于该路由的中间件，保存到 Context 中，依次进行调用
	// 中间件不仅作用在处理流程前，也可以作用在处理流程后，即在用户定义的 Handler 处理完毕后，还可以执行剩下的操作
	handlers []HandlerFunc
	index    int
	engine   *Engine
}

// 获取对应参数的值
func (c *Context) param(key string) string {
	value := c.Params[key]
	return value
}

// NewContext Context的构造方法
func NewContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		//index 是记录当前执行到第几个中间件
		index: -1,
	}
}

//Next 中间件等待用户自己定义的 Handler 处理结束后，再做一些额外的操作，例如计算本次处理所用时间等
//当在中间件中调用Next方法时，控制权交给了下一个中间件，直到调用到最后一个中间件，然后再从后往前，调用每个中间件在Next方法之后定义的部分
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		// 将控制权由当前的中间件交给下一个中间件
		c.handlers[c.index](c)
	}
}
func (c *Context) PostForm(key string) string {
	//FormValue 返回查询的命名组件的第一个值 只包含了 post表单参数
	//c.Req.Form Form属性包含了post表单和url后面跟的get参数
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	//Query 解析 RawQuery 并返回相应的值
	//Get 获取与给定键关联的第一个值。如果没有与键相关联的值，则 Get return the empty string。若要访问多个值，请直接使用 map。
	return c.Req.URL.Query().Get(key)
}

//Status 设置响应头的状态码
func (c *Context) Status(code int) {
	// 在context 中设置状态码
	c.StatusCode = code
	// 写入响应头
	c.Writer.WriteHeader(code)
}

// SetHeader 设置请求头
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

//String 返回的格式为 string 字符串
func (c *Context) String(code int, format string, value ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	// 返回格式字符串 Write 在最后执行 要不之后的SetHeader等等就不能生效
	c.Writer.Write([]byte(fmt.Sprintf(format, value...)))
}

//JSON 设置响应为JSON格式
func (c *Context) JSON(code int, obj interface{}) {
	//设置传输的为JSON格式
	c.SetHeader("Content-Type", "application/json")
	//设置响应头的状态码
	c.Status(code)
	//返回一个写入 w 的新编码器。
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}
