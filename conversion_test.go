//go:build !unsafe

package extnetip

import (
	"net/netip"
	"testing"
)

var (
	mustAddr = netip.MustParseAddr
	mustPfx  = netip.MustParsePrefix
)

var (
	boolSink        bool
	addrSink        addr
	netipAddrSink   netip.Addr
	netipPrefixSink netip.Prefix
)

func TestIdempotent(t *testing.T) {
	t.Parallel()
	v4 := mustAddr("0.0.0.0")
	if wrap(unwrap(v4)) != v4 {
		t.Fatalf("wrap(unwrap(ip)) isn't idempotent, expect: %v, got: %v", v4, wrap(unwrap(v4)))
	}

	v6 := mustAddr("::")
	if wrap(unwrap(v6)) != v6 {
		t.Fatalf("wrap(unwrap(ip)) isn't idempotent, expect: %v, got: %v", v6, wrap(unwrap(v6)))
	}

	v4mappedv6 := mustAddr("::ffff:127.0.0.1")
	if wrap(unwrap(v4mappedv6)) != v4mappedv6 {
		t.Fatalf("wrap(unwrap(ip)) isn't idempotent, expect: %v, got: %v", v4mappedv6, wrap(unwrap(v4mappedv6)))
	}
}

func TestModify(t *testing.T) {
	t.Parallel()
	p4 := unwrap(mustAddr("0.0.0.0"))
	p4.ip.lo++ // add one

	if wrap(p4) != mustAddr("0.0.0.1") {
		t.Fatalf("unwrap -> add one -> wrap not as expected")
	}

	p4.ip.lo-- // sub one
	if wrap(p4) != mustAddr("0.0.0.0") {
		t.Fatalf("unwrap -> sub one -> wrap not as expected")
	}

	// --

	p6 := unwrap(mustAddr("::"))
	p6.ip.lo++ // add one

	if wrap(p6) != mustAddr("::1") {
		t.Fatalf("unwrap -> add one -> wrap not as expected")
	}

	p6.ip.lo-- // sub one
	if wrap(p6) != mustAddr("::") {
		t.Fatalf("unwrap -> sub one -> wrap not as expected")
	}

	// --

	v4mappedv6 := unwrap(mustAddr("::ffff:127.0.0.0"))
	v4mappedv6.ip.lo-- // sub one

	if wrap(v4mappedv6) != mustAddr("::ffff:126.255.255.255") {
		t.Fatalf("unwrap -> add one -> wrap not as expected")
	}

	v4mappedv6.ip.lo++ // add one
	if wrap(v4mappedv6) != mustAddr("::ffff:127.0.0.0") {
		t.Fatalf("unwrap -> sub one -> wrap not as expected")
	}
}
