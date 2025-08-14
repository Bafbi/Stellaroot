package metadata

import (
	"encoding/json"
	"fmt"

	"github.com/bafbi/stellaroot/libs/constant"
)

type Metadata struct {
	Labels      map[string]string `json:"labels,omitempty"`      // User-defined labels for organization/filtering
	Annotations map[string]string `json:"annotations,omitempty"` // System/tool-defined metadata
}

// Label methods
func (m *Metadata) SetLabel(key, value string) {
	if m.Labels == nil {
		m.Labels = make(map[string]string)
	}
	m.Labels[key] = value
}

func (m *Metadata) GetLabel(key string) (string, bool) {
	if m.Labels == nil {
		return "", false
	}
	value, exists := m.Labels[key]
	return value, exists
}

func (m *Metadata) DeleteLabel(key string) {
	if m.Labels != nil {
		delete(m.Labels, key)
	}
}

func (m *Metadata) HasLabel(key, value string) bool {
	if m.Labels == nil {
		return false
	}
	return m.Labels[key] == value
}

// Annotation methods
func (m *Metadata) SetAnnotation(key constant.AnnotationKey, value string) {
	if m.Annotations == nil {
		m.Annotations = make(map[string]string)
	}
	m.Annotations[string(key)] = value
}

func (m *Metadata) GetAnnotation(key constant.AnnotationKey) (string, bool) {
	if m.Annotations == nil {
		return "", false
	}
	value, exists := m.Annotations[string(key)]
	return value, exists
}

func (m *Metadata) DeleteAnnotation(key constant.AnnotationKey) {
	if m.Annotations != nil {
		delete(m.Annotations, string(key))
	}
}

func (m *Metadata) HasAnnotation(key constant.AnnotationKey, value string) bool {
	if m.Annotations == nil {
		return false
	}
	return m.Annotations[string(key)] == value
}

// HasLabels checks if all provided labels match
func (m *Metadata) HasLabels(labels map[string]string) bool {
	if m.Labels == nil {
		return len(labels) == 0
	}
	for key, value := range labels {
		if m.Labels[key] != value {
			return false
		}
	}
	return true
}

// SetStringListAnnotation marshals a []string into a JSON string and sets it as an annotation.
func (m *Metadata) SetStringListAnnotation(key constant.AnnotationKey, list []string) error {
	data, err := json.Marshal(list)
	if err != nil {
		return fmt.Errorf("failed to marshal string list for annotation '%s': %w", key, err)
	}
	m.SetAnnotation(key, string(data))
	return nil
}

// GetStringListAnnotation unmarshals an annotation value from a JSON string into a []string.
func (m *Metadata) GetStringListAnnotation(key constant.AnnotationKey) ([]string, bool, error) {
	val, exists := m.GetAnnotation(key)
	if !exists {
		return nil, false, nil // Not found
	}

	var list []string
	if err := json.Unmarshal([]byte(val), &list); err != nil {
		return nil, true, fmt.Errorf("failed to unmarshal string list from annotation '%s': %w", key, err)
	}
	return list, true, nil // Found and successfully unmarshaled
}

// You can generalize this for any type using interface{} and type assertions/generics (Go 1.18+):
// SetStructuredAnnotation marshals any value into a JSON string and sets it as an annotation.
func (m *Metadata) SetStructuredAnnotation(key constant.AnnotationKey, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal structured data for annotation '%s': %w", key, err)
	}
	m.SetAnnotation(key, string(data))
	return nil
}

// GetStructuredAnnotation unmarshals an annotation value from a JSON string into the target interface.
// 'target' should be a pointer to the type you want to unmarshal into (e.g., &[]string{}, &MyCustomStruct{}).
func (m *Metadata) GetStructuredAnnotation(key constant.AnnotationKey, target any) (bool, error) {
	val, exists := m.GetAnnotation(key)
	if !exists {
		return false, nil // Not found
	}

	if err := json.Unmarshal([]byte(val), target); err != nil {
		return true, fmt.Errorf("failed to unmarshal structured data from annotation '%s': %w", key, err)
	}
	return true, nil // Found and successfully unmarshaled
}

func (m *Metadata) SetBoolAnnotation(key constant.AnnotationKey, value bool) {
	m.SetAnnotation(key, fmt.Sprintf("%t", value))
}

// GetBoolAnnotation retrieves the annotation value associated with the given key
// and attempts to interpret it as a boolean ("true" or "false").
// It returns the boolean value, a boolean indicating if the annotation was present,
// and an error if the value exists but is not a valid boolean string.
// If the annotation does not exist, it returns (false, false, nil).
func (m *Metadata) GetBoolAnnotation(key constant.AnnotationKey) (bool, bool, error) {
	val, exists := m.GetAnnotation(key)
	if !exists {
		return false, false, nil
	}
	switch val {
	case "true":
		return true, true, nil
	case "false":
		return false, true, nil
	default:
		return false, true, fmt.Errorf("annotation '%s' is not a valid bool: %q", key, val)
	}
}
