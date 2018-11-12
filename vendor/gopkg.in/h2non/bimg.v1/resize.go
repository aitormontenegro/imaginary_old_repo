// +build go1.7

package bimg

import (
	"runtime"
    "fmt"
)

// Resize is used to transform a given image as byte buffer
// with the passed options.
func Resize(buf []byte, o Options) ([]byte, error) {
	// Required in order to prevent premature garbage collection. See:
	// https://github.com/h2non/bimg/pull/162
	defer runtime.KeepAlive(buf)

    fmt.Printf("A. size = %d \n", len(buf))
    tt, err := resizer(buf, o)
    fmt.Printf("A1. size = %d \n", len(tt))
    fmt.Printf("AERR. size = %+v \n", err)

	return resizer(buf, o)
}
