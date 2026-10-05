//go:build unsafe

package extnetip

import "net/netip"

// Contains4 reports whether the ip4 address is contained within the pfx.
func Contains4(pfx netip.Prefix, ip4 netip.Addr) bool {
	ip := unwrap(ip4)
	p := unwrap(pfx.Addr())

	bits := pfx.Bits()
	return uint32((ip.ip.lo^p.ip.lo)>>((32-bits)&63)) == 0
}

// Contains6 reports whether the ip6 address is contained within the pfx.
func Contains6(pfx netip.Prefix, ip6 netip.Addr) bool {
	ip := unwrap(ip6)
	p := unwrapPrefix(pfx)

	bits := p.bitsPlusOne - 1
	return ip.ip.xor(p.ip.ip).and(mask6[bits]).isZero()
}
