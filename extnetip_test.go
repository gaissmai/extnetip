// Package extnetip_test contains integration tests and examples
// for the extnetip package.
// These tests verify the correct behavior of IP range and prefix conversion functions.
package extnetip

import (
	"encoding/binary"
	"fmt"
	"iter"
	"math/rand/v2"
	"net/netip"
	"slices"
	"testing"
)

var (
	mpa = netip.MustParseAddr
	mpp = netip.MustParsePrefix
)

type addressFamily uint8

const (
	ip4 addressFamily = iota
	ip6
	ip4in6
)

var ipv4Pfxs = slices.Collect(randPfx(ip4, 16))
var ipv6Pfxs = slices.Collect(randPfx(ip6, 16))

//
// ########### Helper ###########
//

// randIP returns an iterator (iter.Seq) yielding up to n deterministically
// pseudo-random netip.Addr instances for the specified addressFamily.
//
// If n <= 0, the iterator yields nothing. Iteration terminates early if
// the yield function returns false.
func randIP(af addressFamily, n int) iter.Seq[netip.Addr] {
	return func(yield func(netip.Addr) bool) {
		randSrc := rand.New(rand.NewPCG(42, 4242))
		if n <= 0 {
			return
		}

		switch af {
		case ip4:
			for range n {
				var b [4]byte
				binary.BigEndian.PutUint32(b[:], randSrc.Uint32())

				if !yield(netip.AddrFrom4(b)) {
					return
				}
			}

		case ip6:
			for range n {
				var b [16]byte
				binary.BigEndian.PutUint64(b[0:8], randSrc.Uint64())
				binary.BigEndian.PutUint64(b[8:16], randSrc.Uint64())

				if !yield(netip.AddrFrom16(b)) {
					return
				}
			}

		case ip4in6:
			for range n {
				var b [16]byte
				// IPv4-mapped IPv6 header: ::ffff:0:0/96
				b[10] = 0xff
				b[11] = 0xff
				binary.BigEndian.PutUint32(b[12:16], randSrc.Uint32())

				if !yield(netip.AddrFrom16(b)) {
					return
				}
			}
		}
	}
}

// randPfx returns an iterator (iter.Seq) yielding up to n deterministically
// pseudo-random netip.Prefix instances for the specified addressFamily.
//
// Generated prefixes are automatically masked to canonical representation by
// netip.PrefixFrom, zeroing out host bits beyond the prefix length.
//
// If n <= 0, the iterator yields nothing. Iteration terminates early if
// the yield function returns false.
func randPfx(af addressFamily, n int) iter.Seq[netip.Prefix] {
	return func(yield func(netip.Prefix) bool) {
		randSrc := rand.New(rand.NewPCG(42, 4242))
		if n <= 0 {
			return
		}

		for ip := range randIP(af, n) {
			var bits int
			switch af {
			case ip4:
				bits = randSrc.IntN(33) // 0..32
			case ip6:
				bits = randSrc.IntN(129) // 0..128
			case ip4in6:
				bits = 96 + randSrc.IntN(33) // 96..128
			}

			if !yield(netip.PrefixFrom(ip, bits).Masked()) {
				return
			}
		}
	}
}

// assertCoversRange verifies that pfxs is a minimal, contiguous, non-overlapping
// cover of the range [first, last] with no gaps at the boundaries.
func assertCoversRange(first, last netip.Addr, pfxs []netip.Prefix, lenWanted int) error {
	if len(pfxs) != lenWanted {
		return fmt.Errorf("got %d prefix(es), expected %d", len(pfxs), lenWanted)
	}

	// early exit
	if lenWanted == 0 {
		return nil
	}

	// first prefix must start at 'first'
	if f, _ := Range(pfxs[0]); f != first {
		return fmt.Errorf("first prefix starts at %v, want %v", f, first)
	}

	// last prefix must end at 'last'
	if _, l := Range(pfxs[len(pfxs)-1]); l != last {
		return fmt.Errorf("last prefix ends at %v, want %v", l, last)
	}

	// consecutive prefixes must be adjacent (no gaps, no overlaps) with no overlaps
	for i := 1; i < len(pfxs); i++ {

		if pfxs[i-1].Overlaps(pfxs[i]) {
			return fmt.Errorf("%v overlaps %v", pfxs[i-1], pfxs[i])
		}

		_, prevLast := Range(pfxs[i-1])
		curFirst, _ := Range(pfxs[i])

		if prevLast.Next() != curFirst {
			return fmt.Errorf("gap between prefix %v and %v", pfxs[i-1], pfxs[i])
		}
	}

	return nil
}

//
// ########### Tests ###########
//

func TestRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pfx   netip.Prefix
		first netip.Addr
		last  netip.Addr
	}{
		{
			netip.Prefix{},
			netip.Addr{},
			netip.Addr{},
		},
		{
			mpp("0.0.0.0/0"),
			mpa("0.0.0.0"),
			mpa("255.255.255.255"),
		},
		{
			mpp("10.0.0.0/8"),
			mpa("10.0.0.0"),
			mpa("10.255.255.255"),
		},
		{
			mpp("172.16.0.0/12"),
			mpa("172.16.0.0"),
			mpa("172.31.255.255"),
		},
		{
			mpp("::ffff:0.0.0.0/96"),
			mpa("::ffff:0.0.0.0"),
			mpa("::ffff:255.255.255.255"),
		},
		{
			mpp("::/0"),
			mpa("::"),
			mpa("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
		},
		{
			mpp("fe80::/10"),
			mpa("fe80::"),
			mpa("febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
		},
	}

	for n, tt := range tests {
		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			t.Parallel()
			first, last := Range(tt.pfx)
			if first != tt.first {
				t.Errorf("Range(%s), got first: %s, expected: %s", tt.pfx, first, tt.first)
			}
			if last != tt.last {
				t.Errorf("Range(%s), got last: %s, expected: %s", tt.pfx, last, tt.last)
			}
		})
	}
}

func TestPrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip1 netip.Addr
		ip2 netip.Addr
		p   netip.Prefix
		ok  bool
	}{
		{
			netip.Addr{},
			netip.Addr{},
			netip.Prefix{},
			false,
		},
		{
			mpa("0.0.0.0"), // wrong versions
			mpa("::"),
			netip.Prefix{},
			false,
		},
		{
			mpa("0.0.0.1"), // wrong order
			mpa("0.0.0.0"),
			netip.Prefix{},
			false,
		},
		{
			mpa("0.0.0.0"),
			mpa("0.0.0.0"),
			mpp("0.0.0.0/32"),
			true,
		},
		{
			mpa("::"),
			mpa("::"),
			mpp("::/128"),
			true,
		},
		{
			mpa("0.0.0.0"),
			mpa("0.0.0.5"),
			netip.Prefix{},
			false,
		},
		{
			mpa("::"),
			mpa("::5"),
			netip.Prefix{},
			false,
		},
		{
			mpa("0.0.0.0"),
			mpa("0.0.0.3"),
			mpp("0.0.0.0/30"),
			true,
		},
		{
			mpa("0.0.0.0"),
			mpa("255.255.255.255"),
			mpp("0.0.0.0/0"),
			true,
		},
		{
			mpa("10.0.0.0"),
			mpa("10.255.255.255"),
			mpp("10.0.0.0/8"),
			true,
		},
		{
			mpa("172.16.0.0"),
			mpa("172.31.255.255"),
			mpp("172.16.0.0/12"),
			true,
		},
		{
			mpa("::ffff:0.0.0.0"),
			mpa("::ffff:255.255.255.255"),
			mpp("::ffff:0.0.0.0/96"),
			true,
		},
		{
			mpa("::"),
			mpa("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
			mpp("::/0"),
			true,
		},
		{
			mpa("fe80::"),
			mpa("febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
			mpp("fe80::/10"),
			true,
		},
	}

	for n, tt := range tests {
		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			t.Parallel()
			p, ok := Prefix(tt.ip1, tt.ip2)
			if ok != tt.ok {
				t.Errorf("Prefix(%s, %s), got ok: %v, expected: %v", tt.ip1, tt.ip2, ok, tt.ok)
			}
			if p != tt.p {
				t.Errorf("Prefix(%s, %s), got prefix: %s, expected: %s", tt.ip1, tt.ip2, p, tt.p)
			}
		})
	}
}

func TestAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		first   netip.Addr
		last    netip.Addr
		wantLen int // wanted len of prefixes for given range
	}{
		{mpa("0.0.0.0"), mpa("255.255.255.255"), 1},
		{mpa("::"), mpa("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), 1},
		{mpa("::ffff:0.0.0.0"), mpa("::ffff:255.255.255.255"), 1},
		{mpa("10.0.0.0"), mpa("10.255.255.255"), 1},
		{mpa("10.0.0.0"), mpa("10.127.255.255"), 1},
		{mpa("0.0.0.4"), mpa("0.0.0.11"), 2},
		{mpa("10.0.0.0"), mpa("11.10.255.255"), 4},
		{mpa("fe80::"), mpa("fe80::8"), 2},
		{mpa("0.0.0.1"), mpa("255.255.255.254"), 62},
		{mpa("::1"), mpa("ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe"), 254},
	}

	for n, tt := range tests {
		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			t.Parallel()
			var got []netip.Prefix

			// test All()
			for pfx := range All(tt.first, tt.last) {
				got = append(got, pfx)
			}

			if err := assertCoversRange(tt.first, tt.last, got, tt.wantLen); err != nil {
				t.Errorf("All(%s,%s): %v, got %v", tt.first, tt.last, err, got)
			}
		})
	}
}

func TestCommonPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		pfx1   netip.Prefix
		pfx2   netip.Prefix
		expect netip.Prefix // zero value means expect invalid
	}{
		{
			name:   "invalid prefix",
			pfx1:   mpp("192.168.1.0/24"),
			pfx2:   netip.Prefix{},
			expect: netip.Prefix{}, // expect zero
		},
		{
			name:   "different IP versions",
			pfx1:   mpp("192.168.1.0/24"),
			pfx2:   mpp("2001:db8::/32"),
			expect: netip.Prefix{}, // expect zero
		},
		{
			name:   "identical IPv4 prefixes",
			pfx1:   mpp("192.168.1.0/24"),
			pfx2:   mpp("192.168.1.0/24"),
			expect: mpp("192.168.1.0/24"),
		},
		{
			name:   "IPv4 with shorter overlap",
			pfx1:   mpp("192.168.1.0/24"),
			pfx2:   mpp("192.168.2.0/24"),
			expect: mpp("192.168.0.0/22"),
		},
		{
			name:   "IPv4 no overlap",
			pfx1:   mpp("10.0.0.0/8"),
			pfx2:   mpp("192.168.0.0/16"),
			expect: mpp("0.0.0.0/0"),
		},
		{
			name:   "identical IPv6 prefixes",
			pfx1:   mpp("2001:db8::/32"),
			pfx2:   mpp("2001:db8::/32"),
			expect: mpp("2001:db8::/32"),
		},
		{
			name:   "IPv6 with shorter overlap",
			pfx1:   mpp("2001:db8:1::/48"),
			pfx2:   mpp("2001:db8:2::/48"),
			expect: mpp("2001:db8::/46"),
		},
		{
			name:   "IPv4 non-canonical prefix",
			pfx1:   mpp("192.168.1.123/24"), // not masked
			pfx2:   mpp("192.168.1.55/24"),
			expect: mpp("192.168.1.0/24"),
		},
		{
			name:   "IPv6 non-canonical prefix",
			pfx1:   mpp("2001:db8::1/128"),
			pfx2:   mpp("2001:db8:abcd::/32"), // not masked
			expect: mpp("2001:db8::/32"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CommonPrefix(tt.pfx1, tt.pfx2)

			if !tt.expect.IsValid() {
				if got.IsValid() {
					t.Errorf("%s: expected zero prefix, got %v", tt.name, got)
				}
				return
			}

			if got != tt.expect {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expect, got)
			}
		})
	}
}

func TestAll_InvalidInput(t *testing.T) {
	t.Parallel()

	// All of these must return a non-nil, safe-to-call iterator
	// that yields zero elements, without panicking.
	tests := []struct {
		name  string
		first netip.Addr
		last  netip.Addr
	}{
		{"zero addrs", netip.Addr{}, netip.Addr{}},
		{"wrong order", mpa("0.0.0.1"), mpa("0.0.0.0")},
		{"mixed v4/v6", mpa("0.0.0.1"), mpa("::1")},
		{"mixed v4/v4-in-v6", mpa("0.0.0.1"), mpa("::ffff:1.2.3.4")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			it := All(tt.first, tt.last)
			if it == nil {
				t.Fatal("All returned nil iterator, want non-nil zeroIter")
			}
			count := 0
			for range it {
				count++
			}
			if count != 0 {
				t.Errorf("expected 0 elements, got %d", count)
			}
		})
	}
}

func TestIdempotent(t *testing.T) {
	t.Parallel()

	zero := netip.Addr{}

	if wrap(unwrap(zero)) != zero {
		t.Fatalf("wrap(unwrap(zero)) isn't idempotent, expect: %v, got: %v", zero, wrap(unwrap(zero)))
	}

	v4 := mpa("0.0.0.0")
	if wrap(unwrap(v4)) != v4 {
		t.Fatalf("wrap(unwrap(ip)) isn't idempotent, expect: %v, got: %v", v4, wrap(unwrap(v4)))
	}

	v4in6 := mpa("::ffff:127.0.0.1")
	if wrap(unwrap(v4in6)) != v4in6 {
		t.Fatalf("wrap(unwrap(v4in6)) isn't idempotent, expect: %v, got: %v", v4in6, wrap(unwrap(v4in6)))
	}

	v6 := mpa("::")
	if wrap(unwrap(v6)) != v6 {
		t.Fatalf("wrap(unwrap(ip)) isn't idempotent, expect: %v, got: %v", v6, wrap(unwrap(v6)))
	}
}

func TestModify(t *testing.T) {
	t.Parallel()
	p4 := unwrap(mpa("0.0.0.0"))
	p4.ip.lo++ // add one

	if wrap(p4) != mpa("0.0.0.1") {
		t.Fatalf("unwrap -> add one -> wrap not as expected")
	}

	p4.ip.lo-- // sub one
	if wrap(p4) != mpa("0.0.0.0") {
		t.Fatalf("unwrap -> sub one -> wrap not as expected")
	}

	// --

	p6 := unwrap(mpa("::"))
	p6.ip.lo++ // add one

	if wrap(p6) != mpa("::1") {
		t.Fatalf("unwrap -> add one -> wrap not as expected")
	}

	p6.ip.lo-- // sub one
	if wrap(p6) != mpa("::") {
		t.Fatalf("unwrap -> sub one -> wrap not as expected")
	}

	// --

	v4mappedv6 := unwrap(mpa("::ffff:127.0.0.0"))
	v4mappedv6.ip.lo-- // sub one

	if wrap(v4mappedv6) != mpa("::ffff:126.255.255.255") {
		t.Fatalf("unwrap -> add one -> wrap not as expected")
	}

	v4mappedv6.ip.lo++ // add one
	if wrap(v4mappedv6) != mpa("::ffff:127.0.0.0") {
		t.Fatalf("unwrap -> sub one -> wrap not as expected")
	}
}

//
// ########### BenchMarks ###########
//

func BenchmarkRange(b *testing.B) {
	b.Run("v4", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			Range(ipv4Pfxs[i&15])
		}
	})

	b.Run("v6", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			Range(ipv6Pfxs[i&15])
		}
	})
}

func BenchmarkPrefix(b *testing.B) {
	b.Run("v4", func(b *testing.B) {
		firsts := make([]netip.Addr, 0, len(ipv4Pfxs))
		lasts := make([]netip.Addr, 0, len(ipv4Pfxs))
		for _, pfx := range ipv4Pfxs {
			first, last := Range(pfx)
			firsts = append(firsts, first)
			lasts = append(lasts, last)
		}

		for i := 0; b.Loop(); i++ {
			Prefix(firsts[i&15], lasts[i&15])
		}
	})

	b.Run("v6", func(b *testing.B) {
		firsts := make([]netip.Addr, 0, len(ipv6Pfxs))
		lasts := make([]netip.Addr, 0, len(ipv6Pfxs))
		for _, pfx := range ipv6Pfxs {
			first, last := Range(pfx)
			firsts = append(firsts, first)
			lasts = append(lasts, last)
		}

		for i := 0; b.Loop(); i++ {
			Prefix(firsts[i&15], lasts[i&15])
		}
	})

}
