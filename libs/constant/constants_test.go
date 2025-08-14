package constant

import "testing"

func TestPlayerUsernameConstant(t *testing.T) {
	found := false
	for _, k := range AllAnnotationKeys {
		if k == PlayerUsername {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("PlayerUsername not in AllAnnotationKeys")
	}
	if PlayerUsername != "player/username" {
		t.Fatalf("unexpected wire value: %s", PlayerUsername)
	}
}

func TestPlayerOnlineConstant(t *testing.T) {
	found := false
	for _, k := range AllAnnotationKeys {
		if k == PlayerOnline {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("PlayerOnline not in AllAnnotationKeys")
	}
	if PlayerOnline != "player/online" {
		t.Fatalf("unexpected wire value: %s", PlayerOnline)
	}
}

// func TestAnnotationKeysUniqueness(t *testing.T) {
// 	seen := make(map[string]struct{})
// 	for _, k := range AllAnnotationKeys {
// 		if _, exists := seen[k]; exists {
// 			t.Fatalf("duplicate annotation key: %s", k)
// 		}
// 		seen[k] = struct{}{}
// 	}
// }
