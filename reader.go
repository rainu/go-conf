package main

import (
	"io"
	"regexp"
	"slices"
	"strings"
)

var indexRegex = regexp.MustCompile(`^\[[0-9]+\]$`)

type Reader struct {
	r *io.PipeReader
	w *io.PipeWriter

	running bool

	keyValRegex *regexp.Regexp
	options     Options
	args        []string
}

func newReader(args []string, options Options) *Reader {
	r := &Reader{
		args:    args,
		options: options,
	}
	r.r, r.w = io.Pipe()

	r.keyValRegex = regexp.MustCompile(`^` + options.prefix + `([^` + string(options.assignSign) + `]*)` + string(options.assignSign) + `(.*)$`)

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
					r.w.Write([]byte("\""))
					r.w.Write([]byte(segment))
					r.w.Write([]byte("\": "))
					r.w.Write([]byte(l.value))
					r.w.Write([]byte("\n"))
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

		result := r.keyValRegex.FindAllStringSubmatch(key, -1)
		if len(result) != 1 {
			continue
		}

		key = result[0][1]
		value := result[0][2]

		lines = append(lines, line{
			path:  r.splitPreservingBrackets(key),
			value: value,
		})
	}

	// sort lines
	slices.SortFunc(lines, func(a, b line) int {
		return strings.Compare(strings.Join(a.path, ""), strings.Join(b.path, ""))
	})

	return lines
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
