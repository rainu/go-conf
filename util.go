package yacl

// P is a helper function that returns a pointer to the value of type T.
func P[T any](t T) *T {
	return &t
}

// D is a helper function that dereferences a pointer of type T.
// If the pointer is nil, it returns the zero value of type T.
func D[T any](t *T) T {
	if t == nil {
		var zero T
		return zero
	}
	return *t
}
