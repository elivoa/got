package utils

import (
	"sort"
)

func SortStringByLength(strings []string) []string {
	sort.Sort(ByLength(strings))
	return strings
}

// sort by length
type ByLength []string

func (a ByLength) Len() int           { return len(a) }
func (a ByLength) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByLength) Less(i, j int) bool { return len(a[i]) < len(a[j]) }
