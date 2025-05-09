package conf

import (
	"slices"
	"strings"
)

type HelpOptions struct {
	sorter Sorter
	filter Filter
}

func newDefaultHelpOptions() HelpOptions {
	opts := HelpOptions{}

	WithSorter(nil)(&opts)
	WithFilter(nil)(&opts)

	return opts
}

type HelpOption func(*HelpOptions)

// WithSorter sets the sorter for the help output.
func WithSorter(sorter Sorter) HelpOption {
	return func(o *HelpOptions) {
		o.sorter = sorter
	}
}

// WithFilter sets the filter for the help output.
func WithFilter(filter Filter) HelpOption {
	return func(o *HelpOptions) {
		o.filter = filter
	}
}

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

type Filter func(a FieldInfo) bool

func (f *fieldInfos) Filter(filter Filter) FieldInfos {
	f.fi = slices.DeleteFunc(f.fi, func(info fieldInfo) bool {
		return filter(&info)
	})

	return f
}
