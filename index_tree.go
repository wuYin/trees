package trees

type IndexTree interface {
	Insert(key []byte, val interface{})
	Search(key []byte) interface{}
	Delete([]byte) bool
	Size() int
	Dump() map[string]interface{}
}
