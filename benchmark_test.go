package extnetip

import "testing"

//nolint:govet
func BenchmarkConversion(b *testing.B) {
	v4 := mustAddr("0.0.0.0")
	v6 := mustAddr("::")
	addrV4 := peek(v4)
	addrV6 := peek(v6)

	b.Run("peek v4", func(b *testing.B) {
		for b.Loop() {
			addrSink = peek(v4)
		}
	})

	b.Run("peek v6", func(b *testing.B) {
		for b.Loop() {
			addrSink = peek(v6)
		}
	})

	b.Run("back v4", func(b *testing.B) {
		for b.Loop() {
			netipAddrSink = back(addrV4)
		}
	})

	b.Run("back v6", func(b *testing.B) {
		for b.Loop() {
			netipAddrSink = back(addrV6)
		}
	})
}

//nolint:govet
func BenchmarkRange(b *testing.B) {
	v4 := mustPfx("10.1.2.0/24")
	v6 := mustPfx("2001:db8::/56")

	b.Run("v4", func(b *testing.B) {
		for b.Loop() {
			Range(v4)
		}
	})

	b.Run("v6", func(b *testing.B) {
		for b.Loop() {
			Range(v6)
		}
	})
}

//nolint:govet
func BenchmarkPrefix(b *testing.B) {
	first4, last4 := Range(mustPfx("10.1.2.0/24"))
	first6, last6 := Range(mustPfx("2001:db8::/56"))

	b.Run("v4", func(b *testing.B) {
		for b.Loop() {
			Prefix(first4, last4)
		}
	})

	b.Run("v6", func(b *testing.B) {
		for b.Loop() {
			Prefix(first6, last6)
		}
	})
}
