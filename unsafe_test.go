package extnetip

import (
	"net/netip"
	"testing"
	"unsafe"
)

var mustAddr = netip.MustParseAddr

// is it still safe to use unsafe to peek into the internal netip.Addr representation?
func TestSizeof(t *testing.T) {
	s1 := unsafe.Sizeof(addr{})
	s2 := unsafe.Sizeof(netip.Addr{})

	if s1 != s2 {
		t.Fatalf(
			"Address representations differ in size, (%v != %v), maybe internal representation for netip.Addr has changed.",
			s1, s2)
	}
}

func TestIdempotent(t *testing.T) {
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

func TestModify(t *testing.T) {
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
