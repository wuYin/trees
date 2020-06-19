package radix

import (
	"testing"
	"trees/utils"
)

func TestRadix(t *testing.T) {
	var min, max string

	m := make(map[string]interface{})
	tree := NewRadixTree()
	for i, s := range utils.RandStrs(100) {
		k, v := s, s
		if i == 0 || k > max {
			max = k
		}
		if i == 0 || k < min {
			min = k
		}
		m[k] = v
		tree.Insert(k, v)
	}
	if len(m) != tree.Size() {
		t.Fatalf("map size %d, but tree size: %d", len(m), tree.Size())
	}

	for k, v := range m {
		if vv, ok := tree.Get(k); ok && vv != v {
			t.Fatalf("invalid key %q value: %q, want: %q", k, vv, v)
		}
	}

	minKey, _, _ := tree.Min()
	if min != minKey {
		t.Fatalf("want min %q, got %q", min, minKey)
	}
	maxKey, _, _ := tree.Max()
	if max != maxKey {
		t.Fatalf("want max %q, got %q", max, maxKey)
	}
}

func TestInsertAndDelete(t *testing.T) {
	tree := NewRadixTree()
	tree.Insert("romane", 31)
	tree.Insert("roman", 30)
	tree.Delete("roman") // 如果 parent 是 root 节点，则不能删除
}

func TestDelete(t *testing.T) {
	tree := NewRadixTree()
	strs := utils.RandStrs(100)
	m := make(map[string]bool)
	for _, s := range strs {
		tree.Insert(s, nil)
		m[s] = true
	}

	strs = utils.RandStrs(100)
	for _, s := range strs {
		if _, ok := m[s]; !ok {
			m[s] = false
		}
	}
	for k, inserted := range m {
		_, existed := tree.Delete(k)
		if inserted != existed {
			t.Fatalf("delete %s failed, want %t, got %t", k, inserted, existed)
		}
	}
}
