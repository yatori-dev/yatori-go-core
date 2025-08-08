package qutils

import (
	"fmt"
	"testing"
)

// 相似度评分
func TestSimilarity(t *testing.T) {
	target := "hellooooooooooo"
	v1 := "hhlleoooooooooo"
	similarity := Similarity(target, v1)
	fmt.Println(similarity)
}
