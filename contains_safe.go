//go:build !unsafe

package extnetip

import "net/netip"

// silent the linter, isZero is not used under safe builds
var _ = uint128{}.isZero()

// Contains4 reports whether the ip address is contained within the pfx.
func Contains4(pfx netip.Prefix, ip netip.Addr) bool {
	return contains(pfx, ip)
}

// Contains6 reports whether the ip address is contained within the pfx.
func Contains6(pfx netip.Prefix, ip netip.Addr) bool {
	return contains(pfx, ip)
}

func contains(pfx netip.Prefix, ip netip.Addr) bool {
	if ip.Is6() {
		ip = ip.WithZone("")
	}
	return pfx.Contains(ip)
}
