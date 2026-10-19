//go:build !unsafe

// Package extnetip safe conversion implementation.
//
// This file is used when the 'unsafe' build tag is NOT specified.
// It provides safe conversions between netip.Addr and internal uint128
// representation using standard library byte manipulation functions.
//
// While slower than the unsafe version, this approach is suitable for
// environments where unsafe operations are prohibited.

package extnetip

import (
	"encoding/binary"
	"net/netip"
)

// addr holds the IP address as uint128 data and an address family flag.
//
// This struct is used for arithmetic or comparison operations
// on netip.Addr data in a safe manner.
type addr struct {
	ip uint128

	// CRITICAL PERFORMANCE NOTE: Do NOT refactor this to a 'bool'.
	// Using an integer-based address family (addrFamily) allows
	// the Go compiler to keep family flags and conversion states
	// entirely within CPU registers via direct arithmetic and CMOV,
	// preventing forced stack spills and register pressure that
	// would otherwise double the latency of Range() calculations.
	af addrFamily
}

// fromUint128 constructs an addr struct from a uint128 IP representation and af.
func fromUint128(ip uint128, af addrFamily) addr {
	return addr{ip: ip, af: af}
}

// unwrap extracts the internal uint128 bit representation from ip.
//
// Precondition: ip must be a valid address (ip.IsValid() must be true).
// Passing an invalid netip.Addr will trigger a panic.
func unwrap(ip netip.Addr) (a addr) {
	// AsSlice() returns nil if ip is not valid.
	octets := ip.AsSlice()

	if len(octets) == 4 {
		a.af = v4
		a.ip.lo = uint64(binary.BigEndian.Uint32(octets))
		return a
	}

	a.af = v6
	a.ip.hi = binary.BigEndian.Uint64(octets[:8])
	a.ip.lo = binary.BigEndian.Uint64(octets[8:])

	return a
}

// wrap constructs a valid netip.Addr from an internal addr representation
// using the provided scratch buffer to avoid heap allocations.
//
// Precondition: addr must represent a valid IP address. If addr is zero-valued,
// wrap returns netip.IPv6Unspecified() ("::").
func wrap(a addr, buf *[16]byte) (ip netip.Addr) {
	binary.BigEndian.PutUint64(buf[8:], a.ip.lo)
	if a.af == v4 {
		return netip.AddrFrom4([4]byte(buf[12:]))
	}

	binary.BigEndian.PutUint64(buf[:8], a.ip.hi)
	return netip.AddrFrom16(*buf)
}
