package metadata

import (
	"testing"
	"time"

	"github.com/bafbi/stellaroot/libs/constant"
)

func TestBoolDescriptor(t *testing.T) {
	m := &Metadata{}
	desc := constant.NewBoolAnnotationDesc(constant.PlayerOnline)

	// Absent
	if v, ok, err := Get(m, desc); ok || err != nil || v {
		t.Fatalf("expected absent -> ok=false, err=nil, got v=%v ok=%v err=%v", v, ok, err)
	}

	// Set true and read back
	Set(m, desc, true)
	if v, ok, err := Get(m, desc); !ok || err != nil || !v {
		t.Fatalf("roundtrip failed: v=%v ok=%v err=%v", v, ok, err)
	}

	// Corrupt value
	m.SetAnnotation(constant.PlayerOnline, "not-bool")
	if _, ok, err := Get(m, desc); !ok || err == nil {
		t.Fatalf("expected parse error on corrupt value; ok=%v err=%v", ok, err)
	}
}

func TestTimeDescriptor(t *testing.T) {
	m := &Metadata{}
	desc := constant.NewTimeAnnotationDesc(constant.AnnotationKey("player/last_login"), time.RFC3339)
	now := time.Now().UTC().Truncate(time.Second)
	Set(m, desc, now)
	got, ok, err := Get(m, desc)
	if err != nil || !ok || !got.Equal(now) {
		t.Fatalf("time roundtrip mismatch: got=%v ok=%v err=%v", got, ok, err)
	}
}

func TestUUIDDescriptor(t *testing.T) {
	m := &Metadata{}
	desc := constant.NewUUIDAnnotationDesc(constant.AnnotationKey("player/id"))
	Set(m, desc, "123e4567-e89b-12d3-a456-426614174000")
	if v, ok, err := Get(m, desc); !ok || err != nil || v == "" {
		t.Fatalf("uuid roundtrip failed: v=%q ok=%v err=%v", v, ok, err)
	}
	m.SetAnnotation(constant.AnnotationKey("player/id"), "bad-uuid")
	if _, ok, err := Get(m, desc); !ok || err == nil {
		t.Fatalf("expected error for bad uuid")
	}
}
