package utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func LongestPrefix(buf1, buf2 []byte) int {
	max := Min(len(buf1), len(buf2))

	// i 为字节索引，从 0 开始，不满足条件则退出前 i == min(len(s1), len(s2))
	i := 0
	for ; i < max; i++ {
		if buf1[i] != buf2[i] {
			break
		}
	}
	return i
}

func Min(x int, vs ...int) int {
	for _, v := range vs {
		if x > v {
			x = v
		}
	}
	return x
}

var (
	rs    = []rune("abcdefghijklmnopqrstuvwxyz")
	rsLen = len(rs)
)

// 生成 count 个长度在 [minLen, maxLen] 的随机字符串
func RandStrs(count int, minLen, maxLen int) []string {
	strs := make([]string, count)
	for i := 0; i < count; i++ {
		n := 0
		for n < minLen || n > maxLen {
			n = rand.Intn(maxLen)
		}
		strs[i] = RandStr(n)
	}
	return strs
}

func RandStr(length int) string {
	buf := make([]rune, length)
	for i := range buf {
		buf[i] = rs[rand.Intn(rsLen)]
	}
	return string(buf)
}

func Memcpy(dst, src []byte, n int) {
	for i := 0; i < Min(len(dst), len(src), n); i++ {
		dst[i] = src[i]
	}
}

func Memmove(dst, src []byte, n int) {
	for i := 0; i < n; i++ {
		dst[i] = src[i]
	}
}

func ToBytes(s string, base rune) []byte {
	buf := make([]byte, len(s))
	for i, r := range s {
		buf[i] = byte(r - base)
	}
	return buf
}
