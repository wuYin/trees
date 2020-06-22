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
	MIN_NODE4 = 0 // shrink lower limit
	MAX_NODE4 = 4 // expansion upper limit

	MIN_NODE16 = 5
	MAX_NODE16 = 16

	MIN_NODE48 = 17
	MAX_NODE48 = 48

	MIN_NODE256 = 49
	MAX_NODE256 = 256

	MAX_PREFIX_LEN = 8 // line between pessimistic and optimistic search mode
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

type node struct {
	size     int // can't use len(keys) as node current size
	nodeType nodeType

	// internal node
	keys      []byte  // keys
	childs    []*node // typo // key byte -> child *node
	prefix    []byte  // pessimistic, switch to optimistic mode when prefix reach 8 bytes
	prefixLen int     // optimistic, just skip prefixLen partial keys, at last compare whole key in leaf node

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
		keys:     make([]byte, MAX_NODE256), // keys has 256 index
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

//
// utils
//
// node copy
func (n *node) copyMeta(lower *node) {
	n.size = lower.size
	n.prefix = lower.prefix
	n.prefixLen = lower.prefixLen
}

// only use for node48 or node256 which has 256 bytes keys
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

// find child index for key
func (n *node) key2childIndex(k byte) int {
	switch n.nodeType {
	case NODE4, NODE16: // TODO: NODE16 can be implement by Intel SSE instruction
		for i := 0; i < n.size; i++ {
			if n.keys[i] == k {
				return i
			}
		}
		return -1
	case NODE48:
		i := int(n.keys[k]) // convert byte to index int
		if i > 0 {
			return i - 1 // when node16 grow to node48, or insert to node48, index plus a 1 gap, now need reduce
		}
		return -1
	case NODE256:
		return int(k)
	}
	return -1
}

// get match length with other node from index start
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

// get mismatch length with new key
func (n *node) mismatchPrefixLen(key []byte, depth int) int {
	if n.prefixLen <= MAX_PREFIX_LEN {
		// optimistic
		for i := 0; i < n.prefixLen; i++ {
			if key[depth+i] != n.prefix[depth+i] {
				return i
			}
		}
	} else {
		// pessimistic
		i := 0
		for ; i < MAX_PREFIX_LEN; i++ {
			if key[depth+i] != n.prefix[depth+i] {
				return i
			}
		}
		// now continue compare whole key in min node
		leftestLeaf := n.minChild()
		for ; i < n.prefixLen; i++ {
			if key[depth+i] != leftestLeaf.key[depth+i] {
				return i
			}
		}
	}

	return n.prefixLen // sorry, too long to overflow whole prefix
}

// leftest leaf
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
		return n.childs[i-1].minChild() // back 1 byte
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
