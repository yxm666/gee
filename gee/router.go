package gee

import (
	"net/http"
	"strings"
)

type router struct {
	//使用 roots 来存储每种请求方式的 Trie 树根节点 一个请求方法一颗
	roots map[string]*node
	//使用 Handlers 存储每种请求方式的HandlerFunc
	Handlers map[string]HandlerFunc
}

//newRouter 构造函数 初始化Handlers
func newRouter() *router {
	// 使用make 函数初始化 handlers
	return &router{
		roots:    make(map[string]*node),
		Handlers: make(map[string]HandlerFunc),
	}
}

//
func parsePattern(pattern string) []string {
	// 将string 按照 '/' 进行 切分，返回一个[]string
	vs := strings.Split(pattern, "/")
	// 一个长度为0的的string 切片
	parts := make([]string, 0)
	//遍历string 切片
	for _, item := range vs {
		// 如果不是空串 进行添加
		if item != "" {
			parts = append(parts, item)
			// 如果匹配到 * 则返回
			if item[0] == '*' {
				break
			}
		}

	}
	return parts
}

//addRouter 添加路由表
/*对于路由来说，最重要的当然是注册与匹配了。开发服务时，注册路由规则，映射handler；访问时，
匹配路由规则，查找到对应的handler。因此，Trie 树需要支持节点的插入与查询。插入功能很简单，
递归查找每一层的节点，如果没有匹配到当前part的节点，则新建一个，*/
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	//返回一个切层的
	parts := parsePattern(pattern)
	//拼接router表的 key
	key := method + "-" + pattern
	// 每一个请求方法对应一颗前缀树
	_, ok := r.roots[method]
	// 如果不存在方法路由，创建一颗前缀树
	if !ok {
		r.roots[method] = &node{}
	}
	//添加路径到前缀树
	r.roots[method].insert(pattern, parts, 0)
	// 添加到路由表
	r.Handlers[key] = handler
}

func (r router) getRoute(method string, path string) (*node, map[string]string) {
	//分解路径成为 string[]
	searchParts := parsePattern(path)
	// 作为参数
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}
	// 查询获取节点 查询得到的节点肯定是叶子节点 所以有pattern属性
	n := root.search(searchParts, 0)
	// 如果查询的有结果
	if n != nil {
		//获取当前查询节点的解析路径
		parts := parsePattern(n.pattern)
		// 遍历解析路径
		for index, part := range parts {
			// 如果读取的解析路径是通配情况
			if part[0] == ':' {
				//获取参数的值 eg :name -> yxm ====> name:yxm
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				//Join 连接其第一个参数的元素，以创建单个字符串。分隔符字符串分隔符放在结果字符串中的元素之间。
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

//handle 处理请求执行对应的HandlerFunc   将从路由匹配得到的Handler添加到 c.handlers 列表中 执行 c.Next()
func (r *router) handle(c *Context) {
	// 根据字典树 获取字典树node 和 解析后的路径（去掉'/'后的 []string）
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		key := c.Method + "-" + n.pattern
		c.Params = params
		c.handlers = append(c.handlers, r.Handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
