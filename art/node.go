package art

import (
	"bytes"
	"fmt"
	"trees/utils"
)

type nodeType uint8

const (
	NODE4 nodeType = iota
	NODE16
	NODE48
	NODE256
	LEAF
)

const (
	MIN_NODE4 = 0 // 节点收缩下限
	MAX_NODE4 = 4 // 节点膨胀上限

	MIN_NODE16 = 5
	MAX_NODE16 = 16

	MIN_NODE48 = 17
	MAX_NODE48 = 48

	MIN_NODE256 = 49
	MAX_NODE256 = 256

	MAX_PREFIX_LEN = 8 // 当前缀超过 8 bytes 则从悲观模式切换到乐观模式
)

func (n *node) maxSize() (size int) {
	switch n.nodeType {
	case NODE4:
		size = MAX_NODE4
	case NODE16:
		size = MAX_NODE16
	case NODE48:
		size = MAX_NODE48
	case NODE256:
		size = MAX_NODE256
	}
	return size
}

func (n *node) minSize() (size int) {
	switch n.nodeType {
	case NODE4:
		size = MIN_NODE4
	case NODE16:
		size = MIN_NODE16
	case NODE48:
		size = MIN_NODE48
	case NODE256:
		size = MIN_NODE256
	}
	return size
}

type node struct {
	size     int // node 的其他字段均为预分配，其长度不能作为子节点数量
	nodeType nodeType

	// internal node
	keys      []byte  // 有序的子节点 key
	childs    []*node // 指向子节点的指针
	prefix    []byte  // 悲观模式, 为了节省空间，实际只存储一部分公共前缀，最长为 MAX_PREFIX_LEN
	prefixLen int     // 乐观模式，记录完整的前缀长度，比较时找到叶子节点才回溯比较

	// leaf node
	key []byte
	val interface{}
}

func newLeaf(key []byte, val interface{}) *node {
	newKey := make([]byte, len(key))
	copy(newKey, key)
	return &node{
		nodeType: LEAF,
		key:      newKey,
		val:      val,
	}
}

func newNode4() *node {
	return &node{
		nodeType: NODE4,
		keys:     make([]byte, MAX_NODE4),
		childs:   make([]*node, MAX_NODE4),
		prefix:   make([]byte, MAX_PREFIX_LEN),
	}
}

func newNode16() *node {
	return &node{
		nodeType: NODE16,
		keys:     make([]byte, MAX_NODE16),
		childs:   make([]*node, MAX_NODE16),
		prefix:   make([]byte, MAX_PREFIX_LEN),
	}
}

func newNode48() *node {
	return &node{
		nodeType: NODE48,
		keys:     make([]byte, MAX_NODE256), // node48 的 keys 有 256 bytes，查找和空间的折中
		childs:   make([]*node, MAX_NODE48),
		prefix:   make([]byte, MAX_PREFIX_LEN),
	}
}

func newNode256() *node {
	return &node{
		nodeType: NODE256,
		keys:     make([]byte, MAX_NODE256),
		childs:   make([]*node, MAX_NODE256),
		prefix:   make([]byte, MAX_NODE256),
	}
}

func (n *node) isLeaf() bool {
	return n.nodeType == LEAF
}

// check key exactly match leaf node whole key or not
func (n *node) isMatch(key []byte) bool {
	if !n.isLeaf() {
		return false // inner node doesn't storage kv
	}
	return bytes.Compare(n.key, key) == 0
}

func (n *node) isFull() bool {
	return n.size >= n.maxSize()
}

func (n *node) isEmpty() bool {
	return n.size < n.minSize()
}

//
// utils
//
// 从低节点 lower 直接拷贝元信息
func (n *node) copyMeta(lower *node) {
	n.size = lower.size
	n.prefix = lower.prefix
	n.prefixLen = lower.prefixLen
}

// 映射 key 到 child
func (n *node) key2childRef(k byte) **node {
	var emptyNode *node = nil
	if n == nil {
		return &emptyNode
	}
	switch n.nodeType {
	case NODE4, NODE16, NODE48:
		i := n.key2childIndex(k)
		if i == -1 {
			return &emptyNode
		}
		return &n.childs[i]
	case NODE256:
		directChild := n.childs[k]
		if directChild == nil {
			return &emptyNode
		}
		return &directChild
	}
	return &emptyNode
}

// 通过 key 查找 child 的索引位置
func (n *node) key2childIndex(k byte) int {
	switch n.nodeType {
	case NODE4, NODE16:
		for i := 0; i < n.size; i++ { // 只能逐个对比 // TODO: NODE16 SSE 优化查找速度
			if n.keys[i] == k {
				return i
			}
		}
		return -1
	case NODE48:
		i := int(n.keys[k]) // 直接取索引
		if i > 0 {
			return i - 1 // node16 膨胀为 node48 时，所有的 key 都自增了 1，此处需还原 child 真正的索引位置
		}
		return -1
	case NODE256:
		return int(k)
	}
	return -1
}

// 比较与 other 的公共前缀部分的长度
func (n *node) matchPrefixLen(other *node, start int) int {
	end := utils.Min(len(n.key), len(other.key))
	i := start
	for ; i < end; i++ {
		if n.key[i] != other.key[i] {
			return i - start
		}
	}
	return i - start
}

// 与 key 比较，获取第一个不匹配字节在 n.key 中的索引位置
func (n *node) mismatchPrefixLen(key []byte, depth int) int {
	if n.prefixLen <= MAX_PREFIX_LEN {
		// 悲观模式：逐个比较
		for i := 0; i < n.prefixLen; i++ {
			if key[depth+i] != n.prefix[i] {
				return i
			}
		}
	} else {
		i := 0
		for ; i < MAX_PREFIX_LEN; i++ {
			if key[depth+i] != n.prefix[depth+i] {
				return i
			}
		}
		// 切换为乐观模式：取最左叶子节点的完整 key，再逐一比较
		leftestLeaf := n.minChild()
		for ; i < n.prefixLen; i++ {
			if key[depth+i] != leftestLeaf.key[depth+i] {
				return i
			}
		}
	}

	return n.prefixLen // 当前节点的索引完全匹配
}

// 获取最左边的叶子节点，即整棵树的最小 KEY
func (n *node) minChild() *node {
	switch n.nodeType {
	case LEAF:
		return n
	case NODE4, NODE16:
		return n.childs[0].minChild()
	case NODE48:
		i := 0
		for n.childs[i] == nil {
			i++
		}
		return n.childs[i-1].minChild() // 同样要 -1 还原 child 的真实索引
	case NODE256:
		i := 0
		for n.childs[i] == nil {
			i++
		}
		return n.childs[i].minChild()
	default:
		panic(fmt.Sprintf("unknow node type: %d", n.nodeType))
	}
}
