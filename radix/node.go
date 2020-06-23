package radix

import "sort"

type leaf struct {
	key []byte
	val interface{}
}

// 混合了前缀和叶子的节点
// 若 leaf 有值则为叶子节点
// 若 prefix 有值则为前缀节点
// 二者均有值则为混合节点
type node struct {
	leaf   *leaf
	prefix []byte
	edges  edges
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

func (n *node) binSearch(k byte) int {
	l := len(n.edges)
	i := sort.Search(l, func(i int) bool {
		return n.edges[i].k >= k
	})
	if i < l && n.edges[i].k == k {
		return i // 返回前缀边的子节点
	}
	return -1
}

func (n *node) searchEdge(label byte) *node {
	if i := n.binSearch(label); i != -1 {
		return n.edges[i].n // 返回前缀边的子节点
	}
	return nil
}

func (n *node) addEdge(e edge) {
	n.edges = append(n.edges, e)
	n.edges.resort()
}

func (n *node) replaceEdge(k byte, newNode *node) {
	if i := n.binSearch(k); i != -1 {
		n.edges[i].n = newNode
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
	child := n.edges[0].n
	n.prefix = append(n.prefix, child.prefix...)
	n.leaf = child.leaf
	n.edges = child.edges
}

type edge struct {
	k byte  // 边的 byte
	n *node // 末端节点
}

// 方便边的搜索
type edges []edge

func (e edges) Len() int {
	return len(e)
}

func (e edges) Less(i, j int) bool {
	return e[i].k < e[j].k
}

func (e edges) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e edges) resort() {
	sort.Sort(e)
}
