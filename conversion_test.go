package extnetip

import (
	"net/netip"
	"testing"
)

var mustAddr = netip.MustParseAddr

var (
	addrSink      addr
	netipAddrSink netip.Addr
)

func BenchmarkConversion(b *testing.B) {
	v4 := mustAddr("0.0.0.0")
	v6 := mustAddr("::")
	addrV4 := peek(v4)
	addrV6 := peek(v6)

	b.Run("peek v4", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			addrSink = peek(v4)
		}
	})

	b.Run("peek v6", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			addrSink = peek(v6)
		}
	})

	b.Run("back v4", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			netipAddrSink = back(addrV4)
		}
	})

	b.Run("back v6", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			netipAddrSink = back(addrV6)
		}
	})
}

func TestIdempotent(t *testing.T) {
	t.Parallel()
	v4 := mustAddr("0.0.0.0")
	if back(peek(v4)) != v4 {
		t.Fatalf("back(peek(ip)) isn't idempotent, expect: %v, got: %v", v4, back(peek(v4)))
	}

	v6 := mustAddr("::")
	if back(peek(v6)) != v6 {
		t.Fatalf("back(peek(ip)) isn't idempotent, expect: %v, got: %v", v6, back(peek(v6)))
	}

	v4mappedv6 := mustAddr("::ffff:127.0.0.1")
	if back(peek(v4mappedv6)) != v4mappedv6 {
		t.Fatalf("back(peek(ip)) isn't idempotent, expect: %v, got: %v", v4mappedv6, back(peek(v4mappedv6)))
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

	// --

	p6 := peek(mustAddr("::"))
	p6.ip.lo++ // add one

	if back(p6) != mustAddr("::1") {
		t.Fatalf("peek -> add one -> back not as expected")
	}

	p6.ip.lo-- // sub one
	if back(p6) != mustAddr("::") {
		t.Fatalf("peek -> sub one -> back not as expected")
	}

	// --

	v4mappedv6 := peek(mustAddr("::ffff:127.0.0.0"))
	v4mappedv6.ip.lo-- // sub one

	if back(v4mappedv6) != mustAddr("::ffff:126.255.255.255") {
		t.Fatalf("peek -> add one -> back not as expected")
	}

	v4mappedv6.ip.lo++ // add one
	if back(v4mappedv6) != mustAddr("::ffff:127.0.0.0") {
		t.Fatalf("peek -> sub one -> back not as expected")
	}
}
