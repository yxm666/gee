package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

//HandlerFunc 将函数起一个别名 将ResponseWriter Request 封装到Context中
type HandlerFunc func(c *Context)

// Engine implement the interface of ServeHTTP
type (
	RouterGroup struct {
		//前缀 eg：/ /api
		prefix string
		//存储应用在该分组上的中间件
		middlewares []HandlerFunc // support middleware
		//支持分组嵌套 知道当前分组的父亲
		parent *RouterGroup // support nesting
		//还需要有访问 Router 的能力 为了方便，我们可以在Group中，保存一个指针，指向 Engine
		engine *Engine // all groups share a Engine instance
	}

	// Engine 将 Engine 作为最顶层的分组，Engine 拥有 RouterGroup所有的能力
	Engine struct {
		*RouterGroup
		router        *router
		groups        []*RouterGroup     // store all groups
		htmlTemplates *template.Template // 将所有的模版加载进内存
		funcMap       template.FuncMap   // 定义了函数名字符串到函数的映射 所有的自定义模版渲染函数
	}
)

// 实现 handler 接口
func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var middlewares []HandlerFunc
	// 中间件的HandlerFunc加载到Context中  通过URL前缀做一个匹配操作
	for _, group := range e.groups {
		//判断请求适用于哪些中间件 使用URL的前缀进行判断
		if strings.HasPrefix(request.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	// 构造了一个Context对象
	c := NewContext(writer, request)
	c.handlers = middlewares
	c.engine = e
	e.router.handle(c)
}

//New 构造函数
func New() *Engine {
	//使用newRoute 进行初始化 表示对router进行了依赖
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Default use Logger() & Recovery middlewares
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

//Group 建立一个新的RouterGroup
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	// 所有路由都使用同一个引擎
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

//addRoute 添加路径映射 method->pattern : handler
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	// 通过路由组前缀获取路径
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
	//e.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

//Run 封装的http.ListenAndServe
func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

//Use 将中间件应用到某个 Group
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// 解析请求的地址，映射到服务器上文件的真实地址，交给http.FileServer 处理就好了
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {

	//Path.Join 将任意数量的 path 元素连接到一个路径中,用斜线分隔它们。忽略空元素。 eg [a,b,c] --> a/b/c
	absolutePath := path.Join(group.prefix, relativePath)
	// 返回一个hanlder，该handler 会将请求的URL.Path字段中给定前缀 prefix去除后再交由h处理
	// 会向URL.Path 字段中没有给定前缀的请求回复 404
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.param("filepath")
		// 如果文件不存在 或者 无权限进入
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}

}

//Static 暴露给用户，用户可以将磁盘上的某个文件夹 root 映射到路由 relativePath
func (group RouterGroup) Static(relativePath string, root string) {
	// Dir使用限制到指定目录树的本地文件系统实现了http.FileSystem接口。空Dir被视为"."，即代表当前目录。
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}
