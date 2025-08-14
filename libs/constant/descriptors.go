package constant

import (
	"fmt"
	"regexp"
	"time"
)

// AnnotationDescriptor describes how to parse/format a typed annotation value.
// T is the strongly typed representation returned by metadata.Get and accepted by metadata.Set.
type AnnotationDescriptor[T any] struct {
	Key    AnnotationKey
	Parse  func(string) (T, error)
	Format func(T) string
}

// NewStringAnnotationDesc creates a descriptor that leaves values unchanged.
func NewStringAnnotationDesc(key AnnotationKey) AnnotationDescriptor[string] {
	return AnnotationDescriptor[string]{
		Key:    key,
		Parse:  func(s string) (string, error) { return s, nil },
		Format: func(v string) string { return v },
	}
}

// NewBoolAnnotationDesc creates a descriptor for boolean annotations ("true"/"false").
func NewBoolAnnotationDesc(key AnnotationKey) AnnotationDescriptor[bool] {
	return AnnotationDescriptor[bool]{
		Key: key,
		Parse: func(s string) (bool, error) {
			switch s {
			case "true":
				return true, nil
			case "false":
				return false, nil
			default:
				return false, fmt.Errorf("invalid bool: %q", s)
			}
		},
		Format: func(v bool) string {
			if v {
				return "true"
			}
			return "false"
		},
	}
}

// NewTimeAnnotationDesc creates a descriptor for time annotations using a layout (e.g., time.RFC3339).
func NewTimeAnnotationDesc(key AnnotationKey, layout string) AnnotationDescriptor[time.Time] {
	return AnnotationDescriptor[time.Time]{
		Key:    key,
		Parse:  func(s string) (time.Time, error) { return time.Parse(layout, s) },
		Format: func(v time.Time) string { return v.Format(layout) },
	}
}

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// NewUUIDAnnotationDesc validates a UUID string (canonical 36-char with hyphens).
func NewUUIDAnnotationDesc(key AnnotationKey) AnnotationDescriptor[string] {
	return AnnotationDescriptor[string]{
		Key: key,
		Parse: func(s string) (string, error) {
			if uuidRe.MatchString(s) {
				return s, nil
			}
			return "", fmt.Errorf("invalid uuid: %q", s)
		},
		Format: func(v string) string { return v },
	}
}

// NewEnumAnnotationDesc returns a descriptor that admits only values found in allowed.
// The enum type T must be a string-like defined type (constraint ~string).
func NewEnumAnnotationDesc[T ~string](key AnnotationKey, allowed []T) AnnotationDescriptor[T] {
	set := map[string]struct{}{}
	for _, v := range allowed {
		set[string(v)] = struct{}{}
	}
	return AnnotationDescriptor[T]{
		Key: key,
		Parse: func(s string) (T, error) {
			if _, ok := set[s]; ok {
				return T(s), nil
			}
			var zero T
			return zero, fmt.Errorf("invalid enum value: %q", s)
		},
		Format: func(v T) string { return string(v) },
	}
}
