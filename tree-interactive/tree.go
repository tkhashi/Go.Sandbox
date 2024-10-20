package _main

import "strings"

const (
	lastLeft   string = " ┗━"
	centerLeft string = " ┣━"
)

type node struct {
	Value    string
	Children []node
}

type Tree struct {
	nodes []node
}

func (n *node) RenderTree(remainNode []node, indent int) string {
	var b strings.Builder

	for _, node := range remainNode {
		var str string
		if indent > 0 {
			shape := strings.Repeat(" ", (indent-1)*2) + lastLeft + " "
			str += shape
		}

		b.WriteString(str)

		if node.Children != nil {
			childStr := n.RenderTree(node.Children, indent+1)
			b.WriteString(childStr)
		}
	}

	return b.String()
}
