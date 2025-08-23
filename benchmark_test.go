package extnetip

import "testing"

func BenchmarkConversion(b *testing.B) {
	v4 := mustAddr("0.0.0.0")
	v6 := mustAddr("::")
	addrV4 := unwrap(v4)
	addrV6 := unwrap(v6)

	b.Run("unwrap v4", func(b *testing.B) {
		for b.Loop() {
			addrSink = unwrap(v4)
		}
	})

	b.Run("unwrap v6", func(b *testing.B) {
		for b.Loop() {
			addrSink = unwrap(v6)
		}
	})

	b.Run("wrap   v4", func(b *testing.B) {
		for b.Loop() {
			netipAddrSink = wrap(addrV4)
		}
	})

	b.Run("wrap   v6", func(b *testing.B) {
		for b.Loop() {
			netipAddrSink = wrap(addrV6)
		}
	})
}

func BenchmarkRange(b *testing.B) {
	v4 := mustPfx("10.1.2.0/24")
	v6 := mustPfx("2001:db8::/56")

	b.Run("v4", func(b *testing.B) {
		for b.Loop() {
			netipAddrSink, netipAddrSink = Range(v4)
		}
	})

	b.Run("v6", func(b *testing.B) {
		for b.Loop() {
			netipAddrSink, netipAddrSink = Range(v6)
		}
	})
}

func BenchmarkPrefix(b *testing.B) {
	first4, last4 := Range(mustPfx("10.1.2.0/24"))
	first6, last6 := Range(mustPfx("2001:db8::/56"))

	b.Run("v4", func(b *testing.B) {
		for b.Loop() {
			netipPrefixSink, boolSink = Prefix(first4, last4)
		}
	})

	b.Run("v6", func(b *testing.B) {
		for b.Loop() {
			netipPrefixSink, boolSink = Prefix(first6, last6)
		}
	})
}
