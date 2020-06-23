package trees

import (
	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
	"testing"
	"trees/art"
	"trees/radix"
	"trees/utils"
)

func TestIndex(t *testing.T) {
	for _, tree := range []IndexTree{
		art.NewArtTree(),
		radix.NewRadixTree(),
	} {
		m := make(map[string]interface{})
		for _, s := range utils.RandStrs(10, 1, 20) {
			m[s] = s
			tree.Insert([]byte(s), s)
		}
		assert.Equal(t, len(m), tree.Size())

		for k, v := range m {
			assert.Equal(t, v, tree.Search([]byte(k)))
		}
	}
}

func TestArt(t *testing.T) {
	tree := art.NewArtTree()
	tree.Insert([]byte("12345678abcd"), 1)
	tree.Insert([]byte("12345678abef"), 2)
	tree.Insert([]byte("12345678xy"), 3)
	pp.Println(tree.Dump())
}
