package trie

import (
	"strings"
	"testing"
	"trees/utils"
)

func TestTrie(t *testing.T) {
	trie := newTrie()
	m := make(map[string]bool)
	for _, s := range utils.RandStrs(10000) {
		s = strings.ToLower(s)
		v := utils.RandStr(10)
		m[s] = true
		trie.Insert(s, v)
	}
	if len(m) != trie.size {
		t.Fatalf("unexpected trie size: %d, want %d", trie.size, len(m))
	}

	for _, s := range utils.RandStrs(10000) {
		if _, ok := m[s]; !ok {
			m[s] = false
		}
	}

	for s, inserted := range m {
		if _, existed := trie.Delete(s); existed != inserted {
			t.Fatalf("delete %s failed, want %t, got %t", s, inserted, existed)
		}
	}
}

func TestInsert(t *testing.T) {
	trie := newTrie()
	trie.Insert("keyx", "valuex")
	trie.Insert("keyx", "VALUEX")
	v, ok := trie.Get("keyx")
	if !ok {
		t.Fatal("key1 not exist")
	}
	if v.(string) != "VALUEX" {
		t.Fatalf("invalid value:%s", v)
	}
}
