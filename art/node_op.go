package art

import (
	"trees/utils"
)

// 为当前节点添加子节点 newChild，索引 key 是 diffKey
// 如果当前节点已满则膨胀
func (n *node) addChild(diffKey byte, newChild *node) {
	cur := n
	switch cur.nodeType {
	case NODE4, NODE16:
		if !n.isFull() {
			i := n.makeRoomForNewChild(diffKey) // 类比数组的查找插入操作
			cur.keys[i] = diffKey
			cur.childs[i] = newChild
			cur.size++
		} else {
			// 当前节点已满，拷贝数据并膨胀
			n.grow()
			n.addChild(diffKey, newChild)
		}
	case NODE48:
		if !n.isFull() {
			var i int
			var child *node
			for i, child = range n.childs {
				if child == nil {
					break
				}
			}
			n.childs[i] = newChild
			n.keys[diffKey] = byte(i + 1) // 同样要错位
			n.size++
		} else {
			n.grow()
			n.addChild(diffKey, newChild)
		}
	case NODE256:
		if !n.isFull() {
			n.childs[diffKey] = newChild
			n.size++
		}
	}
}

// 节点膨胀
func (n *node) grow() {
	switch n.nodeType {
	// 4 -> 16
	case NODE4:
		next := newNode16()
		next.copyMeta(n)
		for i := 0; i < n.size; i++ { // 直接逐个复制 key 和 child
			next.keys[i] = n.keys[i]
			next.childs[i] = n.childs[i]
		}
		n.replacedBy(next) // cur 会被 GC

	// 16 -> 48
	case NODE16:
		next := newNode48()
		next.copyMeta(n)
		for i, k := range n.keys {
			next.childs[i] = *(n.key2childRef(k))

			// node48 和 node256 一样，都有 256 个 key，但只有 48 childs 指针，不是对应的
			// 这么设计提高了查询速度，也节省了存储空间
			// node48.keys 在初始化时都是 0 值，都会索引到 node48.childs[0] 上
			// 为避免误判，约定将 childs 的索引位置 +1 后再存入 keys，读取时再 -1 即可
			next.keys[k] = byte(i + 1)
		}
		n.replacedBy(next)

	// 48 -> 256
	case NODE48:
		next := newNode256()
		next.copyMeta(n)
		// 逐一复制非空节点
		for _, k := range n.keys {
			next.childs[k] = *(n.key2childRef(k))
		}
		n.replacedBy(next)

	case NODE256:
		panic("node256 needn't grow")
	}
}

// 节点收缩
func (n *node) shrink() {
	switch n.nodeType {
	// 4 -> 1
	case NODE4:
		// 合并唯一子节点
		onlyChild := n.childs[0]

		if onlyChild.isLeaf() { // 唯一子节点为叶子节点，则直接替换
			n.replacedBy(onlyChild)
			return
		}

		// 其他子节点需合并前缀
		if n.prefixLen < MAX_PREFIX_LEN {
			utils.Memcpy(n.prefix[n.prefixLen:], onlyChild.prefix, MAX_PREFIX_LEN)
		}
		n.prefixLen += onlyChild.prefixLen
		n.size = onlyChild.size

		// 替换指向
		n.keys = onlyChild.keys
		n.childs = onlyChild.childs

	// 16 -> 4
	case NODE16:
		prev := newNode4()
		prev.copyMeta(n)
		// 直接逐个替换
		for i := 0; i < MIN_NODE16; i++ {
			prev.keys[i] = n.keys[i]
			prev.childs[i] = n.childs[i]
		}
		n.replacedBy(prev)

	// 48 -> 16
	case NODE48:
		prev := newNode16()
		prev.copyMeta(n)
		childIdx := 0
		for _, k := range n.keys {
			if idx := n.key2childIndex(k); idx != -1 {
				prev.childs[childIdx] = n.childs[idx] // 有序遍历，有序替换
				prev.keys[childIdx] = k
				childIdx++
			}
		}
		n.replacedBy(prev)

	// 256 -> 48
	case NODE256:
		prev := newNode48()
		prev.copyMeta(n)
		childIdx := 0
		for _, k := range n.keys {
			if child := n.childs[n.key2childIndex(k)]; child != nil {
				prev.childs[childIdx] = child
				prev.keys[k] = byte(childIdx + 1) // 依旧自增
				childIdx++
			}
		}
		n.replacedBy(prev)
	}
}

// 替换当前节点
func (n *node) replacedBy(newNode *node) {
	*n = *newNode
}

// 数组查找插入操作
func (n *node) makeRoomForNewChild(diffKey byte) int {
	if n.nodeType != NODE4 && n.nodeType != NODE16 {
		panic("")
	}
	i := 0
	for ; i < n.size; i++ {
		if diffKey < n.keys[i] {
			break
		}
	}

	// TODO: NODE16 SSE 可优化
	for j := n.size; j > i; j-- {
		if n.keys[j-1] > diffKey {
			n.keys[j] = n.keys[j-1]
			n.childs[j] = n.childs[j-1]
		}
	}
	return i
}

// 从内部节点中删除单个 key
func (n *node) delete(k byte) {
	if n.isLeaf() {
		return
	}
	i := n.key2childIndex(k)
	j := 0
	var k2 byte

	switch n.nodeType {
	case NODE4, NODE16: // 删除后 keys 和 childs 必须还对应
		// 删除 child
		// 数组删除中间元素操作
		for ; i < n.size-1; i++ {
			n.childs[i] = n.childs[i+1]
		}
		n.childs[i] = nil

		// 删除 key
		for j, k2 = range n.keys {
			if k == k2 {
				break
			}
		}
		for ; j < n.size-1; j++ {
			n.keys[j] = n.keys[j+1]
		}
		n.keys[j] = byte(0)
	case NODE48:
		// 将 keys 对应置空即可
		childIdx := int(n.keys[i]) - 1
		n.childs[childIdx] = nil
		n.keys[k] = byte(0)
	case NODE256:
		n.childs[int(n.keys[i])] = nil
		n.keys[k] = byte(0)
	}
}
