package conf

import (
	"slices"
	"strings"
)

type Sorter func(a, b FieldInfo) int

func (f *fieldInfos) Sort(sorter Sorter) FieldInfos {
	slices.SortFunc(f.fi, func(a, b fieldInfo) int {
		return sorter(&a, &b)
	})

	return f
}

func PathSorter(a, b FieldInfo) int {
	return strings.Compare(a.Path(), b.Path())
}
