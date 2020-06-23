package art

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
		if n.isFull() {
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
		for i := 0; i < n.size; i++ {
			// 找到空位 j 并复制
			var j int
			var child *node
			for j, child = range n.childs {
				if child == nil {
					break
				}
			}

			next.childs[j] = n.childs[i]
			// node48 和 node256 一样，都有 256 个 key，但只有 48 childs 指针，不是对应的
			// 这么设计提高了查询速度，也节省了存储空间
			// node48.keys 在初始化时都是 0 值，都会索引到 node48.childs[0] 上
			// 为避免误判，约定将 childs 的索引位置 +1 后再存入 keys，读取时再 -1 即可
			next.keys[n.keys[i]] = byte(j + 1)
		}
		n.replacedBy(next)

	// 48 --> 256
	case NODE48:
		next := newNode256()
		next.copyMeta(n)
		// 逐一复制非空节点
		for _, k := range n.keys {
			if child := *(n.key2childRef(k)); child != nil {
				next.childs[k] = child
			}
		}
		n.replacedBy(next)

	case NODE256:
		panic("node256 needn't grow")
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
