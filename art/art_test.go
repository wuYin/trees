package art

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func leaf(key string) *node {
	return newLeaf([]byte(key), nil)
}

func TestCoverage(t *testing.T) {
	tree := NewArtTree()
	tree.Insert([]byte("romanus"), nil)
	t1 := &ArtTree{
		root: leaf("romanus"),
		size: 1,
	}
	assert.True(t, reflect.DeepEqual(*tree, *t1))

	tree.Insert([]byte("romance"), nil)
	t2 := &ArtTree{
		root: &node{
			size:      2,
			nodeType:  NODE4,
			keys:      []byte{'c', 'u', 0x00, 0x00}, // sorted
			childs:    []*node{leaf("romance"), leaf("romanus"), nil, nil},
			prefix:    []byte("roman\x00\x00\x00"),
			prefixLen: 5,
		},
		size: 2,
	}
	assert.True(t, reflect.DeepEqual(*tree, *t2))

	tree.Insert([]byte("romen"), nil)
	t3 := &ArtTree{
		root: &node{
			size:     2,
			nodeType: NODE4,
			keys:     []byte{'a', 'e', 0x00, 0x00},
			childs: []*node{
				&node{
					size:      2,
					nodeType:  NODE4,
					keys:      []byte{'c', 'u', 0x00, 0x00},
					childs:    []*node{leaf("romance"), leaf("romanus"), nil, nil},
					prefix:    []byte("an\x00\x00\x00\x00\x00\x00"),
					prefixLen: 2,
				},
				leaf("romen"),
				nil,
				nil,
			},
			prefix:    []byte("rom\x00\x00\x00\x00\x00"),
			prefixLen: 3,
		},
		size: 3,
	}
	assert.True(t, reflect.DeepEqual(*tree, *t3))
}
