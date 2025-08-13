package constant

import "testing"

func TestPlayerNameConstant(t *testing.T) {
	found := false
	for _, k := range AllAnnotationKeys {
		if k == PlayerName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("PlayerName not in AllAnnotationKeys")
	}
	if PlayerName != "player/name" {
		t.Fatalf("unexpected wire value: %s", PlayerName)
	}
}
