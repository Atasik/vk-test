package main

import "testing"

func BenchmarkCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		сountSource("https://golang.org", "Go")
	}
}
