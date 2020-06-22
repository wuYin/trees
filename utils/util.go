package utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func LongestPrefix(s1, s2 string) int {
	max := Min(len(s1), len(s2))

	// i 为字节索引，从 0 开始，不满足条件则退出前 i == min(len(s1), len(s2))
	i := 0
	for ; i < max; i++ {
		if s1[i] != s2[i] {
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
	minStrLen = 1
	maxStrLen = 10
	rs        = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rsLen     = len(rs)
)

func RandStrs(count int) []string {
	strs := make([]string, count)
	for i := 0; i < count; i++ {
		n := 0
		for n < minStrLen || n > maxStrLen {
			n = rand.Intn(maxStrLen)
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
