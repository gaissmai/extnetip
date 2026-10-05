package extnetip

import (
	"encoding/binary"
	"iter"
	"math/rand/v2"
	"net/netip"
	"slices"
	"testing"
)

type addressFamily uint8

const (
	ip4 addressFamily = iota
	ip6
	ip4in6
)

// RandIP returns an iterator (iter.Seq) yielding up to n deterministically
// pseudo-random netip.Addr instances for the specified addressFamily.
//
// If n <= 0, the iterator yields nothing. Iteration terminates early if
// the yield function returns false.
func RandIP(af addressFamily, n int) iter.Seq[netip.Addr] {
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

// RandPrefix returns an iterator (iter.Seq) yielding up to n deterministically
// pseudo-random netip.Prefix instances for the specified addressFamily.
//
// Generated prefixes are automatically masked to canonical representation by
// netip.PrefixFrom, zeroing out host bits beyond the prefix length.
//
// If n <= 0, the iterator yields nothing. Iteration terminates early if
// the yield function returns false.
func RandPrefix(af addressFamily, n int) iter.Seq[netip.Prefix] {
	return func(yield func(netip.Prefix) bool) {
		randSrc := rand.New(rand.NewPCG(42, 4242))
		if n <= 0 {
			return
		}

		for ip := range RandIP(af, n) {
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

var ipv4Addrs = slices.Collect(RandIP(ip4, 1024))
var ipv6Addrs = slices.Collect(RandIP(ip6, 1024))

var ipv4Pfxs = slices.Collect(RandPrefix(ip4, 1024))
var ipv6Pfxs = slices.Collect(RandPrefix(ip6, 1024))

func BenchmarkAs2xUint64(b *testing.B) {
	b.Run("As2xUint64 v4", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			As2xUint64(ipv4Addrs[i&1023])
		}
	})

	b.Run("As2xUint64 v6", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			As2xUint64(ipv6Addrs[i&1023])
		}
	})

}

func BenchmarkConversion(b *testing.B) {

	b.Run("unwrap v4", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			unwrap(ipv4Addrs[i&1023])
		}
	})

	b.Run("unwrap v6", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			unwrap(ipv6Addrs[i&1023])
		}
	})

	b.Run("wrap   v4", func(b *testing.B) {
		addrs := make([]addr, 0, len(ipv4Addrs))
		for _, ip4 := range ipv4Addrs {
			addrs = append(addrs, unwrap(ip4))
		}

		for i := 0; b.Loop(); i++ {
			wrap(addrs[i&1023])
		}
	})

	b.Run("wrap   v6", func(b *testing.B) {
		addrs := make([]addr, 0, len(ipv6Addrs))
		for _, ip6 := range ipv6Addrs {
			addrs = append(addrs, unwrap(ip6))
		}

		for i := 0; b.Loop(); i++ {
			wrap(addrs[i&1023])
		}
	})
}

func BenchmarkRange(b *testing.B) {
	b.Run("v4", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			Range(ipv4Pfxs[i&1023])
		}
	})

	b.Run("v6", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			Range(ipv6Pfxs[i&1023])
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
			Prefix(firsts[i&1023], lasts[i&1023])
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
			Prefix(firsts[i&1023], lasts[i&1023])
		}
	})

}

func BenchmarkCommonPrefix(b *testing.B) {
	b.Run("v4 mostly short", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			CommonPrefix(ipv4Pfxs[i&1023], ipv4Pfxs[(i^512)&1023])
		}
	})

	b.Run("v4 always long", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			CommonPrefix(ipv4Pfxs[i&1023], ipv4Pfxs[i&1023])
		}
	})

	b.Run("v6 mostly short", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			CommonPrefix(ipv6Pfxs[i&1023], ipv6Pfxs[(i^512)&1023])
		}
	})

	b.Run("v6 always long", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			CommonPrefix(ipv6Pfxs[i&1023], ipv6Pfxs[i&1023])
		}
	})
}
