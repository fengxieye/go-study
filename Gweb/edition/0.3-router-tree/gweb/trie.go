package gweb

import (
	"log"
	"strings"
)

type node struct {
	pattern  string
	part     string
	children []*node
	isWild   bool
}

func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

func (n *node) matchChildren(part string) []*node {
	nodes := []*node{}
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

//func (n *node) insert (pattern string)  {
//	parts := n.parsePattern(pattern)
//	n.add(pattern, parts, 0)
//}
func (n *node) insert(pattern string, parts []string, height int) {
	n.add(pattern, parts, height)
}

func (n *node) parsePattern(pattern string) []string {
	listStr := strings.Split(pattern, "/")
	parts := []string{}
	for _, v := range listStr {
		if v != "" {
			parts = append(parts, v)
			//字符以*开头
			if v[0] == '*' {
				break
			}
		}
	}

	return parts
}

//递归解析
func (n *node) add(pattern string, parts []string, height int) {
	//此时已经没有路径了
	if len(parts) == height {
		//当存在路由覆盖时，让用户解决
		if n.pattern != "" {
			log.Printf("new pattern %s,old pattern %s", pattern, n.pattern)
			panic("路由冲突")
		}
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
		if child.isWild && height == len(parts)-1 {
			//这里处理的例如 /helle 与 /:name ,冲突
			if n.hasEnd() {
				log.Printf(":patter but has end, new pattern %s", pattern)
				panic("路由冲突")
			}
		}
	}
	child.add(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

func (n *node) hasEnd() bool {
	for _, v := range n.children {
		if v.pattern != "" {
			return true
		}
	}

	return false
}
