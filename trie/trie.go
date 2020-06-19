package trie

type TrieTree struct {
	root *node
	size int
}

func newTrie() *TrieTree {
	return &TrieTree{
		root: newNode(false, nil),
		size: 0,
	}
}

type node struct {
	isEnd bool
	val   interface{}
	nexts map[rune]*node // 限制小写字母 key
}

func newNode(isEnd bool, val interface{}) *node {
	return &node{
		val:   val,
		isEnd: isEnd,
		nexts: make(map[rune]*node),
	}
}

func (t *TrieTree) Insert(k string, v interface{}) (interface{}, bool) {
	if !isLower(k) {
		return nil, false
	}

	cur := t.root
	for _, r := range k {
		if next, ok := cur.nexts[r]; ok {
			cur = next
			continue
		}
		if cur.nexts == nil {
			cur.nexts = make(map[rune]*node)
		}
		cur.nexts[r] = newNode(false, nil)
		cur = cur.nexts[r]
	}
	old := cur.val
	cur.val = v
	if cur.isEnd {
		return old, true
	}
	t.size++
	cur.isEnd = true
	return nil, true
}

func (t *TrieTree) Get(k string) (interface{}, bool) {
	if !isLower(k) {
		return nil, false
	}
	cur := t.root
	for _, r := range k {
		next, ok := cur.nexts[r]
		if !ok || next == nil {
			return nil, false
		}
		cur = next
	}
	return cur.val, true
}

func (t *TrieTree) Delete(k string) (interface{}, bool) {
	if !isLower(k) {
		return nil, false
	}
	cur := t.root
	for _, r := range k {
		next, ok := cur.nexts[r]
		if !ok || next == nil {
			return nil, false
		}
		cur = next
	}
	if !cur.isEnd {
		return nil, false
	}

	// 找到 key
	old := cur.val
	cur.val = nil
	// TODO: 回溯向上清理 KEY
	// 思路1：每个 node 记录 parent 地址，回溯判断 len(n.nexts) == 0 则可删除
	// 思路2：daemon 线程定期清理
	// if len(cur.nexts) == 0 && parent != nil {
	// }
	t.size--
	return old, true
}

func (t *TrieTree) Dump() map[string]interface{} {
	var traverse func(s string, n *node, m map[string]interface{})
	traverse = func(s string, n *node, m map[string]interface{}) {
		if n.isEnd {
			m[s] = n.val
		}
		for r, next := range n.nexts {
			if next != nil {
				traverse(s+string(r), next, m)
			}
		}
	}
	m := make(map[string]interface{})
	traverse("", t.root, m)
	return m
}

func (t *TrieTree) Size() int {
	return t.size
}

func isLower(s string) bool {
	for _, r := range s {
		if r-'a' < 0 || r-'a' >= 26 {
			return false
		}
	}
	return true
}
