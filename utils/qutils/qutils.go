package qutils

import "sort"

// 计算两个字符串的Levenshtein距离
func Levenshtein(a, b string) int {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 0; i <= m; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= n; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = min(
				dp[i-1][j]+1,      // 删除
				dp[i][j-1]+1,      // 插入
				dp[i-1][j-1]+cost, // 替换
			)
		}
	}
	return dp[m][n]
}

// 相似度评分
func Similarity(a, b string) float64 {
	maxLen := float64(max(len(a), len(b)))
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(Levenshtein(a, b))/maxLen
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type Co struct {
	index int
	score float64
}

// 相似度匹配并排序
func SimilarityArrayAndSort(target string, v []string) []int {
	coList := make([]Co, len(v))
	for i := 0; i < len(v); i++ {
		coList[i] = Co{index: i, score: Similarity(v[i], target)}
	}
	sort.Slice(coList, func(i, j int) bool {
		return false
	})
	return nil
}

// 直接返回对应最大匹配的ABCD
func SimilarityArraySelect(target string, v []string) string {
	coList := make([]Co, len(v))
	for i := 0; i < len(v); i++ {
		coList[i] = Co{index: i, score: Similarity(v[i], target)}
	}
	var sco = 0.0
	var index = 0
	slist := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "R", "M", "N", "O", "P", "Q"}
	for _, co := range coList {
		if sco < co.score {
			sco = co.score
			index = co.index
		}
	}
	return slist[index]
}

// 直接返回最大匹配的答案文字
func SimilarityArrayAnswer(target string, v []string) string {
	coList := make([]Co, len(v))
	for i := 0; i < len(v); i++ {
		coList[i] = Co{index: i, score: Similarity(v[i], target)}
	}
	var sco = 0.0
	var index = 0
	for _, co := range coList {
		if sco < co.score {
			sco = co.score
			index = co.index
		}
	}
	return v[index]
}
