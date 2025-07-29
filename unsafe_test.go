//go:build unsafe

package extnetip

import (
	"testing"
)

func TestUnsafeIdempotent(t *testing.T) {
	t.Parallel()
	v4 := mustAddr("0.0.0.0")
	if back(peek(v4)) != v4 {
		t.Fatalf("back(peek(ip)) isn't idempotent")
	}

	v6 := mustAddr("::")
	if back(peek(v6)) != v6 {
		t.Fatalf("back(peek(ip)) isn't idempotent")
	}
}

func TestUnsafeModify(t *testing.T) {
	t.Parallel()
	p4 := peek(mustAddr("0.0.0.0"))
	p4.ip.lo++ // add one

	if back(p4) != mustAddr("0.0.0.1") {
		t.Fatalf("peek -> add one -> back not as expected")
	}

	p4.ip.lo-- // sub one
	if back(p4) != mustAddr("0.0.0.0") {
		t.Fatalf("peek -> sub one -> back not as expected")
	}

	p6 := peek(mustAddr("::"))
	p6.ip.lo++ // add one

	if back(p6) != mustAddr("::1") {
		t.Fatalf("peek -> add one -> back not as expected")
	}

	p6.ip.lo-- // sub one
	if back(p6) != mustAddr("::") {
		t.Fatalf("peek -> sub one -> back not as expected")
	}
}
