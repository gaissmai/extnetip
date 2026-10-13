//go:build unsafe

package extnetip

import (
	"net/netip"
	"testing"
)

var mpa = netip.MustParseAddr

func TestIdempotent(t *testing.T) {
	t.Parallel()
	v4 := mpa("0.0.0.0")
	if wrap(unwrap(v4)) != v4 {
		t.Fatalf("wrap(unwrap(ip)) isn't idempotent, expect: %v, got: %v", v4, wrap(unwrap(v4)))
	}

	v6 := mpa("::")
	if wrap(unwrap(v6)) != v6 {
		t.Fatalf("wrap(unwrap(ip)) isn't idempotent, expect: %v, got: %v", v6, wrap(unwrap(v6)))
	}

	v4mappedv6 := mpa("::ffff:127.0.0.1")
	if wrap(unwrap(v4mappedv6)) != v4mappedv6 {
		t.Fatalf("wrap(unwrap(ip)) isn't idempotent, expect: %v, got: %v", v4mappedv6, wrap(unwrap(v4mappedv6)))
	}
}

func TestAs2xUint64_Idempotent(t *testing.T) {
	t.Parallel()
	// AddrFrom2xUint64(As2xUint64(ipv4)) is not idempotent, AddrFrom2xUint64 always returns an IPv6 address.

	v6 := mpa("2001:db8::1")
	if AddrFrom2xUint64(As2xUint64(v6)) != v6 {
		t.Fatalf("AddrFrom2xUint64(As2xUint64(v6)) isn't idempotent, expect: %v, got: %v", v6, AddrFrom2xUint64(As2xUint64(v6)))
	}

	ip4in6 := mpa("::ffff:127.0.0.1")
	if AddrFrom2xUint64(As2xUint64(ip4in6)) != ip4in6 {
		t.Fatalf("AddrFrom2xUint64(As2xUint64(ip4in6)) isn't idempotent, expect: %v, got: %v", ip4in6, AddrFrom2xUint64(As2xUint64(ip4in6)))
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

func TestContainsUnsafe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pfx  netip.Prefix
		ip   netip.Addr
		want bool
	}{
		{
			name: "IPv4 match inside prefix",
			pfx:  netip.MustParsePrefix("192.168.1.0/24"),
			ip:   netip.MustParseAddr("192.168.1.50"),
			want: true,
		},
		{
			name: "IPv4 match exact network address",
			pfx:  netip.MustParsePrefix("192.168.1.0/24"),
			ip:   netip.MustParseAddr("192.168.1.0"),
			want: true,
		},
		{
			name: "IPv4 no match outside prefix",
			pfx:  netip.MustParsePrefix("192.168.1.0/24"),
			ip:   netip.MustParseAddr("192.168.2.1"),
			want: false,
		},
		{
			name: "IPv4 slash zero prefix",
			pfx:  netip.MustParsePrefix("0.0.0.0/0"),
			ip:   netip.MustParseAddr("8.8.8.8"),
			want: true,
		},
		{
			name: "IPv6 match inside prefix",
			pfx:  netip.MustParsePrefix("2001:db8::/32"),
			ip:   netip.MustParseAddr("2001:db8::1234"),
			want: true,
		},
		{
			name: "IPv6 no match outside prefix",
			pfx:  netip.MustParsePrefix("2001:db8::/32"),
			ip:   netip.MustParseAddr("2001:db9::1234"),
			want: false,
		},
		{
			name: "IPv6 with zone identifier (Unlike standard Contains, ContainsUnsafe accepts zones)",
			pfx:  netip.MustParsePrefix("2001:db8::/32"),
			ip:   netip.MustParseAddr("2001:db8::1%eth0"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ContainsUnsafe(tt.pfx, tt.ip)
			if got != tt.want {
				t.Errorf("ContainsUnsafe(%v, %v) = %v, want %v", tt.pfx, tt.ip, got, tt.want)
			}
		})
	}
}

func BenchmarkContainsUnsafe(b *testing.B) {
	b.Run("v4", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			ContainsUnsafe(ipv4Pfxs[i&1023], ipv4Addrs[i&1023])
		}
	})

	b.Run("v6", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			ContainsUnsafe(ipv6Pfxs[i&1023], ipv6Addrs[i&1023])
		}
	})
}
