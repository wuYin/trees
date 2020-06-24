package art

import (
	"trees/utils"
)

type ArtTree struct {
	root *node
	size int
}

// 创建空树
func NewArtTree() *ArtTree {
	return &ArtTree{root: nil, size: 0}
}

func (t *ArtTree) Insert(key []byte, val interface{}) {
	key = appendNULL(key)
	t.insert(t.root, &t.root, 0, key, val)
}

// 递归遍历 key 直到遇到叶子节点
// 处理 lazy expansion 和 mismatch
func (t *ArtTree) insert(cur *node, curRef **node, depth int, key []byte, val interface{}) {
	// 1. 空树或空叶子节点
	if cur == nil {
		*curRef = newLeaf(key, val)
		t.size++
		return
	}

	// 2. 处理叶子节点的 lazy expansion
	if cur.isLeaf() {
		// 2.1. key 存在则先返回
		if cur.isMatch(key) {
			return
		}

		// 2.2. 当前节点会被公共前缀父节点替换掉，当前节点切割公共前缀后，与新叶子节点一起连接到该父节点
		leaf := newLeaf(key, val)
		commonLen := cur.matchPrefixLen(leaf, depth)

		parent := newNode4()
		parent.prefixLen = commonLen // 当前深度的公共前缀长度
		utils.Memcpy(parent.prefix, key[depth:depth+commonLen], utils.Min(parent.prefixLen, MAX_PREFIX_LEN))

		// 节点替换，用第一个字节作为 key 建立 childs 指针
		*curRef = parent
		parent.addChild(cur.key[depth+commonLen], cur)
		parent.addChild(key[depth+commonLen], leaf)

		t.size++
		return
	}

	// 3. 处理内部节点的分裂
	diffIdx := cur.mismatchPrefixLen(key, depth)
	if diffIdx != cur.prefixLen {
		parent := newNode4() // 分裂父节点

		// 添加叶子节点
		leaf := newLeaf(key, val)
		parent.addChild(key[depth+diffIdx], leaf)

		// 添加当前节点

		// 拷贝前缀到父节点
		parent.prefixLen = diffIdx // 注意此处 index 和 len 的关系是相等的
		utils.Memcpy(parent.prefix, cur.prefix, diffIdx)

		if cur.prefixLen < MAX_PREFIX_LEN {
			// 在当前节点的部分前缀匹配成功
			cur.prefixLen -= diffIdx + 1 // 1: diffKey
			parent.addChild(cur.prefix[diffIdx], cur)
			utils.Memmove(cur.prefix, cur.prefix[(diffIdx+1):], cur.prefixLen) // 之后 prefixLen 和 prefix 是同步的
		} else {
			cur.prefixLen -= diffIdx + 1
			// 从子节点拿完整的 key 来做前缀匹配
			leftestLeaf := cur.minChild()
			parent.addChild(leftestLeaf.key[depth+diffIdx], cur)
			utils.Memmove(cur.prefix, leftestLeaf.key[depth+diffIdx+1:], utils.Min(cur.prefixLen, MAX_PREFIX_LEN))
		}

		*curRef = parent
		t.size++
		return
	}

	// 4. 处理一般情况：跳过当前内部节点，继续下沉寻找目标叶子节点
	depth += cur.prefixLen
	next := cur.key2childRef(key[depth])
	if *next == nil {
		// 找到叶子节点的目标位置
		cur.addChild(key[depth], newLeaf(key, val))
		t.size++
		return
	}

	// 继续下沉
	t.insert(*next, next, depth+1, key, val)
}

func (t *ArtTree) Search(key []byte) interface{} {
	key = appendNULL(key)
	return t.search(t.root, key, 0)
}

func (t *ArtTree) search(n *node, key []byte, depth int) interface{} {
	if n == nil {
		return nil
	}
	if n.isLeaf() {
		if n.isMatch(key) {
			return n.val
		}
		return nil
	}
	diffIdx := n.mismatchPrefixLen(key, depth)
	// 在 n 节点内部不匹配
	if diffIdx != n.prefixLen {
		return nil
	}

	depth += n.prefixLen
	next := n.key2childRef(key[depth])
	return t.search(*next, key, depth+1)
}

func (t *ArtTree) Delete(key []byte) bool {
	key = appendNULL(key)
	return t.delete(t.root, nil, 0, key)
}

func (t *ArtTree) delete(cur *node, parent *node, depth int, key []byte) bool {
	// search leaf node and delete it
	if cur == nil {
		return false
	}

	diffIdx := cur.mismatchPrefixLen(key, depth)
	if diffIdx != cur.prefixLen {
		return false
	}

	if cur.isLeaf() {
		if !cur.isMatch(key) {
			return false
		}
		if parent == nil {
			return false // TODO: delete root
		}
		// 1. delete the leaf
		leafPrefixKey := key[depth]
		parent.delete(leafPrefixKey)
		parent.size--
		t.size--

		// 2. lazy expansion
		if parent.size == 1 {

		}

		return true
	}

	depth += cur.prefixLen
	next := cur.key2childRef(key[depth])
	return t.delete(*next, cur, depth, key) // depth no need +1, depth is index now
}

func (t *ArtTree) Size() int {
	return t.size
}

func (t *ArtTree) Dump() map[string]interface{} {
	var traverse func(n *node, m map[string]interface{})
	traverse = func(n *node, m map[string]interface{}) {
		if n == nil {
			return
		}
		if n.isLeaf() {
			m[string(n.key)] = n.val
		}
		for _, k := range n.keys {
			child := *(n.key2childRef(k))
			if child != nil {
				traverse(child, m)
			}
		}
	}

	m := make(map[string]interface{})
	traverse(t.root, m)
	return m
}
