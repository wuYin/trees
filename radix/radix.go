package radix

import (
	"strings"
	"trees/utils"
)

// 基数树
type RadixTree struct {
	root *node
	size int
}

func NewRadixTree() *RadixTree {
	return &RadixTree{
		root: &node{}, // root 始终空
		size: 0,
	}
}

// 新增：创建新的 edge 和 leaf
// 更新：已作为前缀节点存在，更新值并返回
//
// 分三种情况
// roman   --> romane // 创建为父节点 `roman`
// romanus --> romane // 创建 `roman` 为父前缀节点
// romanex --> romane // 创建新叶子子节点 `x`
//
func (t *RadixTree) Insert(k string, v interface{}) interface{} {
	originKey := k
	var parent *node
	cur := t.root

	// 向下分裂前缀并插入节点
	for {

		// 修改 root 节点的值
		// 或是切割后发现是前缀节点，则更新值或添加叶子节点
		if len(k) == 0 {
			if cur.isLeafNode() {
				old := cur.leaf.v
				cur.leaf.v = v
				return old
			}
			cur.leaf = &leaf{k: originKey, v: v}
			t.size++
			return nil
		}

		parent = cur
		cur = cur.searchEdge(k[0])

		// 1. 没有边则创建
		if cur == nil { // root --r--> [prefix:`roman`, k:`romane`]
			e := edge{
				label: k[0],
				node: &node{
					leaf: &leaf{
						k: originKey,
						v: v,
					},
					prefix: k,
					edges:  nil,
				},
			}
			parent.addEdge(e) // 记录到父节点
			t.size++
			return nil
		}

		// 2. 有子节点，先找出最长前缀
		commonLen := utils.LongestPrefix(cur.prefix, k)
		if commonLen == len(cur.prefix) {
			// 2.1. 节点前缀被完全覆盖，则切割后继续向下走
			// `romance` --> roman --> `romane`
			// 			 		   --> `romanus`
			k = k[commonLen:]
			continue
		}

		// 2.2. 不覆盖节点前缀，则生成父节点，按第一个异构字节分裂出 2 条边，连接到 2 个子节点（一旧一新）
		// `romane`
		// `roman` --e--> `romane`
		// 		   --u--> `romanus`
		commonNode := &node{
			prefix: k[:commonLen],
		}
		parent.replaceEdge(k[0], commonNode) // `r`

		// 将当前节点挪到子节点
		commonNode.addEdge(edge{
			label: cur.prefix[commonLen],
			node:  cur,
		})
		cur.prefix = cur.prefix[commonLen:] // `romane` --> `e`

		// 创建新节点
		newLeaf := &leaf{k: originKey, v: v}
		k = k[commonLen:]
		if len(k) == 0 { // k 恰好是分裂的前缀，则创建当前叶子节点 // `roman` --> `romane`
			commonNode.leaf = newLeaf
			t.size++
			return nil
		}

		commonNode.addEdge(edge{
			label: k[0],
			node: &node{
				leaf:   newLeaf,
				prefix: k,
			},
		})
		t.size++
		return nil
	}
}

// 删除
func (t *RadixTree) Delete(k string) (interface{}, bool) {

	var parent *node
	var label byte
	cur := t.root

	// 1. 查找 k 节点
	for {
		if len(k) == 0 { // `us`
			if !cur.isLeafNode() {
				return nil, false // 必须要是叶子节点，避免误删
			}
			break // bingo
		}

		parent = cur
		label = k[0]
		cur = cur.searchEdge(label) // `r` `u`
		if cur == nil {
			return nil, false // 边不存在
		}

		if !strings.HasPrefix(k, cur.prefix) {
			return nil, false // 边存在，但节点不存在
		}

		// 继续向下查找
		k = k[len(cur.prefix):] // `romanus` - `roman` = `us`
	}

	// 2. 删除叶子节点
	old := cur.leaf.v
	cur.leaf = nil
	t.size--

	switch len(cur.edges) {
	case 0:
		// 2.1. 当前节点不是前缀节点，先删除边
		if parent != nil { // 删的不是 root
			parent.deleteEdge(label) // `u`
		}
	case 1:
		// 2.2. 叶子被清理，若只有一个子节点则合并
		cur.replaceByOnlyChild()
	}

	// 3. 若父节点为前缀节点，且只有一个节点，也要合并
	if parent != nil && !parent.isLeafNode() && len(parent.edges) == 1 {
		if parent != t.root { // 根节点不能被替换，之前的 insert 等操作都是从 root 出发
			parent.replaceByOnlyChild()
		}
	}
	return old, true
}

// 查找
func (t *RadixTree) Get(k string) (interface{}, bool) {
	cur := t.root
	for {
		if len(k) == 0 {
			if !cur.isLeafNode() {
				return nil, false
			}
			return cur.leaf.v, true
		}

		cur = cur.searchEdge(k[0])
		if cur == nil {
			return nil, false
		}
		if !strings.HasPrefix(k, cur.prefix) {
			return nil, false
		}
		k = k[len(cur.prefix):]
	}
}

func (t *RadixTree) Dump() map[string]interface{} {
	var traverse func(n *node, m map[string]interface{})
	traverse = func(n *node, m map[string]interface{}) {
		if n == nil {
			return
		}
		if n.isLeafNode() {
			m[n.leaf.k] = n.leaf.v
		}
		for _, e := range n.edges {
			traverse(e.node, m)
		}
	}
	m := make(map[string]interface{})
	traverse(t.root, m)
	return m
}

func (t *RadixTree) Size() int {
	return t.size
}

func (t *RadixTree) Min() (string, interface{}, bool) {
	cur := t.root
	for {
		if cur.isLeafNode() {
			return cur.leaf.k, cur.leaf.v, true
		}
		if cur.isPrefixNode() {
			cur = cur.edges[0].node
			continue
		}
		return "", nil, false
	}
}

func (t *RadixTree) Max() (string, interface{}, bool) {
	cur := t.root
	for {
		if cur.isPrefixNode() {
			cur = cur.edges[len(cur.edges)-1].node
			continue
		}
		if cur.isLeafNode() {
			return cur.leaf.k, cur.leaf.v, true
		}
		return "", nil, false
	}
}
