package main

import (
	"io"
	"testing"
	"testing/iotest"
)

func BenchmarkCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := iotest.ErrReader(io.EOF)
		Count(r, "Go")
	}

}
