package art

import (
	"bytes"
)

func cp(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

// 对于 insert 和 search 的递归操作，若下沉前目标 key 已遍历完毕，再取 diffKey 时会直接数组溢出
// 论文的 C 实现，类型为 char* 的 key 尾部都有 \0，下沉时遍历结束依旧可以再取到 \0 的diffKey，和任何有值的 key 比较都不相等，遍历直接结束
// Go 的实现也需要模拟尾部的空后缀，防止下沉溢出
func appendNULL(key []byte) []byte {
	if bytes.IndexByte(key, 0x00) > 0 {
		return key
	}
	return append(key, 0x00)
}
