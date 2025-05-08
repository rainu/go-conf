package conf

import (
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

var indexRegex = regexp.MustCompile(`^\[[0-9]+\]$`)
var onlyNumberRegex = regexp.MustCompile(`^[0-9]+$`)

type Reader struct {
	r *io.PipeReader
	w *io.PipeWriter

	running bool

	reKeyVal          *regexp.Regexp
	reKeyValFlag      *regexp.Regexp
	reKeyValShort     *regexp.Regexp
	reKeyValShortFlag *regexp.Regexp

	options    Options
	fieldInfos *fieldInfos
	args       []string
}

func newReader(args []string, dst *fieldInfos, options Options) *Reader {
	r := &Reader{
		args:       args,
		fieldInfos: dst,
		options:    options,
	}
	r.r, r.w = io.Pipe()

	r.reKeyVal = regexp.MustCompile(`^` + options.prefixLong + `([^` + string(options.assignSign) + `]*)` + string(options.assignSign) + `(.*)$`)
	r.reKeyValShort = regexp.MustCompile(`^` + options.prefixShort + `([^` + string(options.assignSign) + `]*)` + string(options.assignSign) + `(.*)$`)

	r.reKeyValFlag = regexp.MustCompile(`^` + options.prefixLong + `([^` + string(options.assignSign) + `]*)$`)
	r.reKeyValShortFlag = regexp.MustCompile(`^` + options.prefixShort + `([^` + string(options.assignSign) + `]*)$`)

	return r
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if !r.running {
		go r.run()
		r.running = true
	}
	return r.r.Read(p)
}

func (r *Reader) Close() error {
	return r.w.Close()
}

func (r *Reader) run() {
	defer r.w.Close()

	lines := r.collectLines()
	indentation := make(map[string]bool)
	for _, l := range lines {
		for i, segment := range l.path {
			indent := strings.Repeat("  ", i)

			key := strings.Join(l.path[:i+1], ".")
			if !indentation[key] {
				if i == len(l.path)-1 {
					// "<indent><segment>: <value>"
					r.w.Write([]byte(indent))
					segment = strings.TrimPrefix(segment, "[")
					segment = strings.TrimSuffix(segment, "]")

					if onlyNumberRegex.MatchString(segment) {
						// this is a primitive array value
						r.w.Write([]byte("- "))
						r.w.Write([]byte(l.value))
						r.w.Write([]byte("\n"))
					} else {
						r.w.Write([]byte("\""))
						r.w.Write([]byte(segment))
						r.w.Write([]byte("\": "))
						r.w.Write([]byte(l.value))
						r.w.Write([]byte("\n"))
					}
				} else {
					if strings.HasPrefix(segment, "[") && indexRegex.MatchString(segment) {
						// "<indent>-"  (Array-Element)
						r.w.Write([]byte(indent))
						r.w.Write([]byte("-\n"))
					} else {
						// "<indent><segment>:"
						r.w.Write([]byte(indent))
						segment = strings.TrimPrefix(segment, "[")
						segment = strings.TrimSuffix(segment, "]")
						r.w.Write([]byte("\""))
						r.w.Write([]byte(segment))
						r.w.Write([]byte("\":\n"))
					}
				}
				indentation[key] = true
			}
		}
	}
}

type line struct {
	path  []string
	value string
}

func (r *Reader) collectLines() []line {
	lines := make([]line, 0, len(r.args))

	for i := 0; i < len(r.args); i += 1 {
		key := r.args[i]
		var value string

		if !r.tryShort(key, &key, &value) {
			if !r.tryShortFlag(key, &key, &value) {
				if !r.tryLong(key, &key, &value) {
					if !r.tryLongFlag(key, &key, &value) {
						continue
					}
				}
			}
		}

		path := r.splitPreservingBrackets(key)
		if r.fieldInfos != nil {
			lastNode := path[len(path)-1]
			if !strings.HasSuffix(lastNode, "]") {
				// check the type of the corresponding field
				// it could be a slice ...
				info := r.fieldInfos.findByPath(path)
				if info != nil && strings.HasPrefix(info.Type, "[]") {
					// this is a slice, but the argument has no index given
					// so here we add the index
					// "i" is not the correct index,
					// but is only necessary that the index is an increasing number
					// (because of the sorting later)
					path = append(path, fmt.Sprintf("[%d]", i))
				}
			}

		}

		lines = append(lines, line{
			path:  path,
			value: value,
		})
	}

	// sort lines
	slices.SortFunc(lines, func(a, b line) int {
		return strings.Compare(strings.Join(a.path, ""), strings.Join(b.path, ""))
	})

	return lines
}

func (r *Reader) tryLong(line string, key, value *string) bool {
	result := r.reKeyVal.FindAllStringSubmatch(line, -1)
	if len(result) == 1 {
		*key = result[0][1]
		*value = result[0][2]
		return true
	}
	return false
}

func (r *Reader) tryLongFlag(line string, key, value *string) bool {
	result := r.reKeyValFlag.FindAllStringSubmatch(line, -1)
	if len(result) == 1 {
		*key = result[0][1]
		*value = "true"
		return true
	}
	return false
}

func (r *Reader) tryShort(line string, key, value *string) bool {
	// for short variant we need the fieldInfos
	if r.fieldInfos == nil {
		return false
	}

	result := r.reKeyValShort.FindAllStringSubmatch(line, -1)
	if len(result) == 1 {
		k := result[0][1]

		// search for fieldInfo with given short key
		corProperty := r.fieldInfos.findByShort(k)
		if corProperty == nil {
			return false
		}

		// convert to long-variant and delegate to the long-variant
		line = r.options.prefixLong + corProperty.Path.key(r.options, "0") + string(r.options.assignSign) + result[0][2]
		return r.tryLong(line, key, value)
	}
	return false
}

func (r *Reader) tryShortFlag(line string, key, value *string) bool {
	// for short variant we need the fieldInfos
	if r.fieldInfos == nil {
		return false
	}

	result := r.reKeyValShortFlag.FindAllStringSubmatch(line, -1)
	if len(result) == 1 {
		k := result[0][1]

		// search for fieldInfo with given short key
		corProperty := r.fieldInfos.findByShort(k)
		if corProperty == nil {
			return false
		}

		// convert to long-variant and delegate to the long-variant
		line = r.options.prefixLong + corProperty.Path.key(r.options, "0")
		return r.tryLongFlag(line, key, value)
	}
	return false
}

func (r *Reader) splitPreservingBrackets(s string) []string {
	var result []string
	var current strings.Builder
	inBrackets := false

	for i := 0; i < len(s); i++ {
		switch rune(s[i]) {
		case '[':
			inBrackets = true
			current.WriteByte(s[i])
		case ']':
			inBrackets = false
			current.WriteByte(s[i])
		case r.options.keyDelimiter:
			if inBrackets {
				current.WriteByte(s[i])
			} else if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(s[i])
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
