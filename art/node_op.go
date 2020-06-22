package art

// add node for cur child, and with partial key
// grow node if key size reach upper node limit
func (n *node) addChild(diffKey byte, newChild *node) {
	cur := n
	switch cur.nodeType {
	case NODE4, NODE16:
		if !n.isFull() {
			i := n.makeRoomForNewChild(diffKey)
			cur.keys[i] = diffKey
			cur.childs[i] = newChild
			cur.size++
		} else {
			// cur node4 is full, need to upgrade to node16 and insert again
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
			n.keys[diffKey] = byte(i + 1) // skip index 0
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

// grow cur to next bigger size node
func (n *node) grow() {
	switch n.nodeType {
	// 4 -> 16
	case NODE4:
		next := newNode16()
		next.copyMeta(n)
		for i := 0; i < n.size; i++ { // copy keys and childs directly
			next.keys[i] = n.keys[i]
			next.childs[i] = n.childs[i]
		}
		n.replacedBy(next)

	// 16 -> 48
	case NODE16:
		next := newNode48()
		next.copyMeta(n)
		for i := 0; i < n.size; i++ {
			// find a empty index j in next.childs for child
			var j int
			var child *node
			for j, child = range n.childs {
				if child == nil {
					break
				}
			}

			next.childs[j] = n.childs[i]
			// node48 has 256 keys but 48 children, they two are not corresponding
			// NOTICE
			// node48.keys initialized as 256 zero value bytes, it means all 256 keys pointing to node48.childs[0]
			// so we can't save child to node48.childs[0] directly, we need
			// there is no uint8 overflow, node48.childs upper limited to 48, so j+1 in [1, 49], still less than 256
			next.keys[n.keys[i]] = byte(j + 1)
		}
		n.replacedBy(next)

	// 48 --> 256
	case NODE48:
		next := newNode256()
		next.copyMeta(n)
		// copy children corresponding with keys, so they two will be sorted
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

// replace current node
func (n *node) replacedBy(newNode *node) {
	*n = *newNode
}

// make room for newNode in node4 or node16
func (n *node) makeRoomForNewChild(diffKey byte) int {
	if n.nodeType != NODE4 && n.nodeType != NODE16 {
		panic("")
	}
	// 1. find diffKey index in cur.keys, so get child index in cur.childs
	i := 0
	for ; i < n.size; i++ {
		if diffKey < n.keys[i] {
			break
		}
	}

	// 2. move childs[i:] backward to make one position for newChild
	// TODO: NODE16 can be optimize by SSE
	for j := n.size; j > i; j-- {
		if n.keys[j-1] > diffKey {
			n.keys[j] = n.keys[j-1]
			n.childs[j] = n.childs[j-1]
		}
	}
	return i
}
