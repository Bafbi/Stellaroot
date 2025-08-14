package metadata

import (
	"github.com/bafbi/stellaroot/libs/constant"
)

// Get retrieves and parses an annotation using the descriptor.
// Returns (zero, false, nil) when the annotation is not present.
// Returns (zero, true, err) when present but invalid.
func Get[T any](m *Metadata, d constant.AnnotationDescriptor[T]) (T, bool, error) {
	var zero T
	if m == nil {
		return zero, false, nil
	}
	raw, ok := m.GetAnnotation(d.Key)
	if !ok {
		return zero, false, nil
	}
	v, err := d.Parse(raw)
	if err != nil {
		return zero, true, err
	}
	return v, true, nil
}

// Set formats and stores an annotation using the descriptor.
func Set[T any](m *Metadata, d constant.AnnotationDescriptor[T], v T) {
	if m == nil {
		return
	}
	m.SetAnnotation(d.Key, d.Format(v))
}
