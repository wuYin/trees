package radix

import (
	"bytes"
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

// 新增或更新
func (t *RadixTree) Insert(key []byte, val interface{}) {
	originKey := make([]byte, len(key))
	copy(originKey, key)
	newLeaf := &leaf{key: originKey, val: val}

	var parent *node
	cur := t.root

	for {
		// 3. 找到目标叶子节点，插入或更新值
		// 修改 root 节点的值，或切割后发现是前缀节点，则更新值或添加叶子节点
		if len(key) == 0 {
			if cur.isLeafNode() {
				cur.leaf.val = val
				return
			}
			cur.leaf = newLeaf
			t.size++
			return
		}

		parent = cur
		cur = cur.searchEdge(key[0])

		// 1. 没有边指向叶子节点的边则创建
		if cur == nil {
			e := edge{
				k: key[0],
				n: &node{
					leaf:   newLeaf,
					prefix: key,
					edges:  nil,
				},
			}
			parent.addEdge(e) // 记录到父节点
			t.size++
			return
		}

		// 2. 有子节点则分裂
		commonLen := utils.LongestPrefix(cur.prefix, key)
		if commonLen == len(cur.prefix) {
			// 2.1. 当前节点的前缀被完全覆盖，则分裂后继续下沉
			key = key[commonLen:]
			continue
		}

		// 2.2. 不覆盖则分裂当前节点
		commonNode := &node{
			prefix: key[:commonLen],
		}
		parent.replaceEdge(key[0], commonNode) // 变更指向到新父节点
		commonNode.addEdge(edge{
			k: cur.prefix[commonLen],
			n: cur, // 将当前节点挪到子节点
		})
		cur.prefix = cur.prefix[commonLen:] // 切割前缀

		key = key[commonLen:]
		if len(key) == 0 { // key 恰好是分裂出的前缀，则 commonNode 为混合节点
			commonNode.leaf = newLeaf
			t.size++
			return
		}

		commonNode.addEdge(edge{
			k: key[0],
			n: &node{
				prefix: key,
				leaf:   newLeaf,
			},
		})
		t.size++
		return
	}
}

// 删除
func (t *RadixTree) Delete(key []byte) bool {
	var parent *node
	var k byte
	cur := t.root

	// 1. 查找 key 对应的叶子节点
	for {
		if len(key) == 0 {
			if !cur.isLeafNode() {
				return false // 必须是叶子节点，避免删除
			}
			break // bingo
		}

		parent = cur
		k = key[0]
		cur = cur.searchEdge(k)
		if cur == nil {
			return false // 边不存在
		}

		if !bytes.HasPrefix(key, cur.prefix) {
			return false // 边存在，但节点不存在
		}

		// 切割前缀，继续向下查找
		key = key[len(cur.prefix):]
	}

	// 2. 删除叶子节点
	cur.leaf = nil
	t.size--

	switch len(cur.edges) {
	case 0:
		// 2.1. 当前节点只是叶子节点，先删除边
		if parent != nil { // 删 root 不用删边
			parent.deleteEdge(k)
		}
	case 1:
		// 2.2. 当前节点是混合节点，且只有一个子节点，删除后要上浮该子节点
		cur.replaceByOnlyChild()
	}

	// 2.3. 若父节点只是前缀节点，且只有一个子节点，要继续上浮
	if parent != nil && !parent.isLeafNode() && len(parent.edges) == 1 {
		if parent != t.root { // 根节点不能被替换，之前的 insert 等操作都是从 root 出发
			parent.replaceByOnlyChild()
		}
	}
	return true
}

// 查找
func (t *RadixTree) Search(key []byte) interface{} {
	cur := t.root
	for {
		if len(key) == 0 {
			if !cur.isLeafNode() {
				return nil
			}
			return cur.leaf.val
		}

		cur = cur.searchEdge(key[0])
		if cur == nil {
			return nil
		}
		if !bytes.HasPrefix(key, cur.prefix) {
			return nil
		}
		key = key[len(cur.prefix):]
	}
}

func (t *RadixTree) Dump() map[string]interface{} {
	var traverse func(n *node, m map[string]interface{})
	traverse = func(n *node, m map[string]interface{}) {
		if n == nil {
			return
		}
		if n.isLeafNode() {
			m[string(n.leaf.key)] = n.leaf.val
		}
		for _, e := range n.edges {
			traverse(e.n, m)
		}
	}
	m := make(map[string]interface{})
	traverse(t.root, m)
	return m
}

func (t *RadixTree) Size() int {
	return t.size
}

func (t *RadixTree) Min() ([]byte, interface{}) {
	cur := t.root
	for {
		if cur.isLeafNode() {
			return cur.leaf.key, cur.leaf.val
		}
		if cur.isPrefixNode() {
			cur = cur.edges[0].n
			continue
		}
		return nil, nil
	}
}

func (t *RadixTree) Max() ([]byte, interface{}) {
	cur := t.root
	for {
		if cur.isPrefixNode() {
			cur = cur.edges[len(cur.edges)-1].n
			continue
		}
		if cur.isLeafNode() {
			return cur.leaf.key, cur.leaf.val
		}
		return nil, nil
	}
}
