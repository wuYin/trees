package art

import (
	"trees/utils"
)

type ArtTree struct {
	root *node
	size int
}

// create an empty tree
func NewArtTree() *ArtTree {
	return &ArtTree{root: nil, size: 0}
}

func (t *ArtTree) Insert(key []byte, val interface{}) {
	t.insert(t.root, &t.root, 0, key, val)
}

// traverse key recursively until find a leaf node to insert
// handle lazy expansion and mismatch
func (t *ArtTree) insert(cur *node, curRef **node, depth int, key []byte, val interface{}) {
	// 1. handle empty tree or new leaf node
	if cur == nil {
		*curRef = newLeaf(key, val)
		t.size++
		return
	}

	// 2. handle leaf lazy expansion
	// `romanus` --> leaf `romance`
	if cur.isLeaf() {
		// 2.1. key existed, just return  TODO: need update
		if cur.isMatch(key) {
			return
		}

		// 2.2. cur replaced by a new inner node storing the existing and new leaf
		// compare cur key with new key, get common prefix length as new inner node prefixLen
		leaf := newLeaf(key, val)
		commonLen := cur.matchPrefixLen(leaf, depth)

		parent := newNode4()
		parent.prefixLen = commonLen // whole prefix's length
		// if commonLen > MAX_PREFIX_LEN, then switch to optimistic mode while reach the node
		utils.Memcpy(parent.prefix, key[depth:], utils.Min(commonLen, MAX_PREFIX_LEN)) // partial prefix

		// replace cur with new node and add the 2 children
		*curRef = parent
		parent.addChild(cur.key[depth+commonLen], cur)
		parent.addChild(key[depth+commonLen], leaf)

		t.size++
		return
	}

	// 3. handle inner node mismatch
	if cur.prefixLen != 0 {
		mismatch := cur.mismatchPrefixLen(key, depth)

		if mismatch != cur.prefixLen {
			parent := newNode4() // same split flow
			parent.prefixLen = mismatch
			utils.Memcpy(parent.prefix, cur.prefix, mismatch) // FIXME: mismatch may be overflow cur.prefix

			if cur.prefixLen < MAX_PREFIX_LEN {
				// 3.1. mismatch still in cur.prefix
				parent.addChild(cur.prefix[mismatch], cur)
				cur.prefixLen -= mismatch
				utils.Memmove(cur.prefix, cur.prefix[mismatch:], utils.Min(cur.prefixLen, MAX_PREFIX_LEN))
			} else {
				// 3.2. mismatch overflow cur.prefix to minChild.prefix
				parent.addChild(cur.prefix[depth+mismatch], cur)
				cur.prefixLen -= mismatch
				leftestLeaf := cur.minChild() // TODO
				utils.Memmove(cur.prefix, leftestLeaf.key[depth+mismatch:], utils.Min(cur.prefixLen, MAX_PREFIX_LEN))
			}
			// now just cut off mismatched prefix from current node
			for i := range cur.prefix {
				if i >= cur.prefixLen {
					cur.prefix[i] = 0x00
				}
			}

			// leaf node as child
			leaf := newLeaf(key, val)
			parent.addChild(key[depth+mismatch], leaf)
			*curRef = parent

			t.size++
			return
		}
	}

	// 4. handle normal case: perfectly match prefix and carry on
	depth += cur.prefixLen
	next := cur.key2childRef(key[depth])
	if *next == nil {
		// bingo, found a leaf
		cur.addChild(key[depth], newLeaf(key, val))
		return
	}

	// carry on
	t.insert(*next, next, depth+1, key, val)
}
