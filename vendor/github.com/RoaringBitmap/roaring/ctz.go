//go:build go1.9
// +build go1.9

// "go1.9", from Go version 1.9 onward
// See https://golang.org/pkg/go/build/#hdr-Build_Constraints

package roaring

import "math/bits"

func countTrailingZeros(x uint64) int {
	return bits.TrailingZeros64(x)
}
