//
// 基数树实现
// https://en.wikipedia.org/wiki/Radix_tree
//
package radix

import "sort"

//
// 存放实际 kv 的叶子节点
//
type leaf struct {
	k string
	v interface{}
}

//
// 包含前缀和前缀边的树节点
//
type node struct {
	leaf   *leaf  // 是叶子节点则有值
	prefix string // 当前节点抽离出的子节点的前缀
	edges  edges  // 边
}

func (n *node) isLeafNode() bool {
	return n.leaf != nil
}

func (n *node) isPrefixNode() bool {
	return len(n.edges) > 0
}

func (n *node) isMixedNode() bool {
	return n.isLeafNode() && n.isPrefixNode()
}

func (n *node) binSearch(label byte) int {
	l := len(n.edges)
	i := sort.Search(l, func(i int) bool {
		return n.edges[i].label >= label
	})
	if i < l && n.edges[i].label == label {
		return i // 返回前缀边的子节点
	}
	return -1
}

func (n *node) searchEdge(label byte) *node {
	if i := n.binSearch(label); i != -1 {
		return n.edges[i].node // 返回前缀边的子节点
	}
	return nil
}

func (n *node) addEdge(e edge) {
	n.edges = append(n.edges, e)
	n.edges.resort()
}

func (n *node) replaceEdge(label byte, newNode *node) {
	if i := n.binSearch(label); i != -1 {
		n.edges[i].node = newNode
		return
	}
	panic("replace unexpected")
}

func (n *node) deleteEdge(label byte) {
	if i := n.binSearch(label); i != -1 {
		// 保持有序
		copy(n.edges[i:], n.edges[i+1:])
		n.edges[len(n.edges)-1] = edge{}
		n.edges = n.edges[:len(n.edges)-1]
		return
	}
	panic("delete unexpected")
}

// 提升唯一子节点
func (n *node) replaceByOnlyChild() {
	child := n.edges[0].node
	n.prefix += child.prefix
	n.leaf = child.leaf
	n.edges = child.edges
}

//
// 前缀边
//
type edge struct {
	label byte  // 边值
	node  *node // 末端节点
}

// 方便边的搜索
type edges []edge

func (e edges) Len() int {
	return len(e)
}

func (e edges) Less(i, j int) bool {
	return e[i].label < e[j].label
}

func (e edges) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e edges) resort() {
	sort.Sort(e)
}
