package radix

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"trees/utils"
)

func TestRadix(t *testing.T) {
	var min, max string

	m := make(map[string]interface{})
	tree := NewRadixTree()
	for i, s := range utils.RandStrs(100, 1, 10) {
		k, v := s, s
		if i == 0 || k > max {
			max = k
		}
		if i == 0 || k < min {
			min = k
		}
		m[k] = v
		tree.Insert([]byte(k), v)
	}
	assert.Equal(t, len(m), tree.size)

	for k, v := range m {
		assert.Equal(t, v, tree.Search([]byte(k)))
	}

	minKey, _ := tree.Min()
	assert.Equal(t, min, string(minKey))
	maxKey, _ := tree.Max()
	assert.Equal(t, max, string(maxKey))
}

func TestInsertAndDelete(t *testing.T) {
	tree := NewRadixTree()
	tree.Insert([]byte("romane"), 31)
	tree.Insert([]byte("roman"), 30)
	tree.Delete([]byte("roman")) // 如果 parent 是 root 节点，则不能删除
}

func TestDelete(t *testing.T) {
	tree := NewRadixTree()
	strs := utils.RandStrs(100, 1, 10)
	m := make(map[string]bool)
	for _, s := range strs {
		tree.Insert([]byte(s), nil)
		m[s] = true
	}

	strs = utils.RandStrs(100, 1, 10)
	for _, s := range strs {
		if _, ok := m[s]; !ok {
			m[s] = false
		}
	}
	for k, inserted := range m {
		_, existed := tree.Delete([]byte(k))
		assert.Equal(t, inserted, existed)
	}
}
