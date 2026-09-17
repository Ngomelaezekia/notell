package main

import "strings"

func userIDValue(value string) uint64 {
	var out uint64
	for _, r := range strings.TrimSpace(value) {
		if r < '0' || r > '9' {
			return 0
		}
		out = out*10 + uint64(r-'0')
	}
	return out
}
