package common

import (
	"fmt"
	"testing"
)

func TestBytesCombine(t *testing.T) {
	fmt.Println(BytesCombine([]byte("one"), []byte("two"), []byte("测试")))
}

func TestSha256(t *testing.T) {
	t.Log(Sha256("hello world"))
}

func TestGetCurrentPath(t *testing.T) {
	t.Log(GetCurrentPath())
}
