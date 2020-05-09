package msv

import (
	"path"
	"strings"
)

// ShiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func ShiftPathN(p string, n int) (head, tail string) {
	tail = p
	for i := 0; i < n; i++ {
		head, tail = ShiftPath(tail)
	}
	return head, tail
}
