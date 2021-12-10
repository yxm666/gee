package gee

import "strings"

/*与普通的树不同，为了实现动态路由匹配，加上了isWild这个参数。
即当我们匹配 /p/go/doc/这个路由时，第一层节点，p精准匹配到了p，
第二层节点，go模糊匹配到:lang，那么将会把lang这个参数赋值为go，继续下一层匹配。
我们将匹配的逻辑，包装为一个辅助函数。*/
type node struct {
	// 待匹配路由 例如 /p/:lang
	pattern string
	// 路由中的一部分，例如:lang
	part string
	// 子节点 例如[doc,tutorial,intro]
	children []*node
	// 是否精确匹配 part 含有 : 或 * 时为true
	isWild bool
}

//matchChild 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	//遍历当前节点的所有子节点
	for _, child := range n.children {
		// 如果部分路径匹配 或者 是模糊匹配
		if child.part == part || child.isWild {
			return child
		}
	}
	// 没有匹配的就返回nil
	return nil
}

//matchChildren 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	//用来返回当前的节点的所有匹配孩子
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// DFS方法
//insert 递归查找每一层的节点，如果没有匹配到当前 part 的节点，则新建一个
//pattern 是指的访问路径  parts 是切分后的string切片
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		// 只有叶子节点才会有 pattern 变量 所以在结束的时候 判断节点的pattern if n.pattern == "" 则没有匹配
		n.pattern = pattern
		return
	}
	//按照从低到高 取出 part 切片中的当前递归层次处理的部分
	part := parts[height]
	// 寻找第一个匹配成功的节点
	child := n.matchChild(part)

	// 如果没找到匹配的节点
	if child == nil {
		// 如果没有匹配到一个节点，那么构造一个节点，part isWild 是根据当前part的开头是否匹配进行判断
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		// 将child添加到当前节点的孩子列表中
		n.children = append(n.children, child)
	}
	// 递归调用
	child.insert(pattern, parts, height+1)
}

//search 查询功能，同样也是递归查询每一层的节点，退出规则是，匹配到了 *，匹配失败，或者匹配到了第 len(parts)层节点
func (n *node) search(parts []string, height int) *node {
	//string 是否以 prefix 为前缀
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		// 匹配到最后匹配失败了
		if n.pattern == "" {
			return nil
		}
		return n
	}
	// 获取当前层次的part
	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		// 递归遍历
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}
