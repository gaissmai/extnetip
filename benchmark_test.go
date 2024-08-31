package extnetip

import "testing"

func BenchmarkConversion(b *testing.B) {
	v4 := mustAddr("0.0.0.0")
	v6 := mustAddr("::")
	addrV4 := peek(v4)
	addrV6 := peek(v6)

	b.Run("peek v4", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			addrSink = peek(v4)
		}
	})

	b.Run("peek v6", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			addrSink = peek(v6)
		}
	})

	b.Run("back v4", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			netipAddrSink = back(addrV4)
		}
	})

	b.Run("back v6", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			netipAddrSink = back(addrV6)
		}
	})
}
