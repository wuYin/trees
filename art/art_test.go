package art

import (
	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
	"testing"
	"trees/utils"
)

//
// 功能 case
//
func TestArtBasic(t *testing.T) {
	// insert
	t1 := NewArtTree()
	t1.Insert([]byte("ab"), "AB")
	assert.Equal(t, t1.Size(), 1)
	assert.Equal(t, t1.root.size, 0)
	assert.Equal(t, t1.root.nodeType, LEAF)
	assert.Equal(t, t1.root.key, []byte{'a', 'b', 0x00}) // 尾部要加上空字节
	assert.Equal(t, t1.Search([]byte("ab")), "AB")

	// search
	assert.Equal(t, t1.Search([]byte("ab")), "AB")

	// split
	// LEAF -> NODE4
	t1.Insert([]byte("ac"), "AC")
	assert.Nil(t, t1.Search([]byte("a"))) // 内部节点
	assert.Equal(t, t1.Search([]byte("ab")), "AB")
	assert.Equal(t, t1.Search([]byte("ac")), "AC")
}

//
// 膨胀 case
//
func TestNodeExpansion(t *testing.T) {
	// [l,r) 插入范围 kv
	insertRange := func(t *ArtTree, l, r int) {
		for i := l; i < r; i++ {
			t.Insert([]byte{byte(i)}, byte(i))
		}
	}
	t1 := NewArtTree()
	assert.Nil(t, t1.root)

	// NULL -> LEAF
	insertRange(t1, 0, 1)
	assert.Equal(t, t1.root.nodeType, LEAF)

	// LEAF -> NODE4
	insertRange(t1, 1, 2)
	assert.Equal(t, t1.root.nodeType, NODE4)
	insertRange(t1, 2, 4)
	assert.True(t, t1.root.isFull())

	// NODE4 -> NODE16
	insertRange(t1, 4, 5)
	assert.Equal(t, t1.root.nodeType, NODE16)
	insertRange(t1, 5, 16)
	assert.True(t, t1.root.isFull())

	// NODE16 -> NODE48
	insertRange(t1, 16, 17)
	assert.Equal(t, t1.root.nodeType, NODE48)
	insertRange(t1, 17, 48)
	assert.True(t, t1.root.isFull())

	// NODE48 -> NODE256
	insertRange(t1, 48, 49)
	assert.Equal(t, t1.root.nodeType, NODE256)
	insertRange(t1, 49, 255)
	assert.True(t, t1.root.isFull())
	pp.Println(t1.root.keys)
}

//
// 大量 key case
//
func TestSearch(t *testing.T) {
	tree := NewArtTree()
	m := make(map[string]interface{})
	for _, s := range utils.RandStrs(10000, 1, 10) {
		m[s] = s
		tree.Insert([]byte(s), s)
	}
	pp.Println(len(m))
	assert.Equal(t, len(m), tree.Size()) // 验证去重

	for k, v := range m {
		assert.Equal(t, v, tree.Search([]byte(k)))
	}
}

func TestFailCase(t *testing.T) {
	tree := NewArtTree()
	for _, s := range []string{"a", "jia", "injqlsc", "wwcjvh", "xrclo", "ipyq", "ugp", "cio", "xy", "wzswkb",
		"sdosx", "eqjximo", "ovzgh", "xfbfzyntz", "mq", "gkixrt", "eu", "ocnmxinya", "ffylfl", "j", "oljmrox", "eyv",
		"joqebuw", "krtlnca", "tjzq", "evqtyvtu", "wurd", "b", "oyjh", "juezo", "ywq"} {
		tree.Insert([]byte(s), s)
	}
	pp.Println(tree.Search([]byte("tjzq")))
}
