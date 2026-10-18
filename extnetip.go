// Package extnetip is an extension to net/netip providing auxiliary functions
// for converting IP prefixes to IP ranges and vice versa.
//
// The calculations are done efficiently in uint128 space, avoiding conversions
// to/from byte slices. These extensions allow easy implementation of third-party
// libraries for IP range management on top of net/netip.
//
// This package supports both safe and unsafe modes of operation. When built with
// the 'unsafe' build tag, conversions use unsafe.Pointer for better performance.
// Without the tag, safe byte-slice based conversions are used.
package extnetip

import (
	"iter"
	"net/netip"
)

// empty iterator for wrong input parameter
var zeroIter iter.Seq[netip.Prefix] = func(func(netip.Prefix) bool) {}

// addrFamily represents the internal address family classification
// used to distinguish between IPv4 and IPv6 representations.
type addrFamily uint

const (
	v4 addrFamily = 4
	v6 addrFamily = 6
)

// Range returns the inclusive IP address range [first, last]
// covered by the given prefix p.
//
// If p is invalid, Range returns zero values.
func Range(p netip.Prefix) (first, last netip.Addr) {
	if !p.IsValid() {
		return
	}

	ip := p.Addr()
	bits := p.Bits()

	// Check address family from the value directly while it resides in CPU registers.
	// NOTE: ip.Is4() must be called BEFORE unwrap(ip) because taking the address
	// of `ip` via unwrap forces the compiler to allocate a physical stack slot,
	// inhibiting register promotion and downstream compiler optimizations.
	// Extract internal representation of the prefix address as an addr struct
	// only after value-based inspection is complete.
	af := v4
	if !ip.Is4() {
		af = v6
	}
	pa := unwrap(ip)

	if af == v4 {
		// IPv4 addresses are embedded in IPv6 space with a 96-bit prefix
		bits += 96
	}
	mask := mask6[bits] // get the network mask as uint128

	// Calculate first IP in range: ip & mask
	buf := [16]byte{}

	first128 := pa.ip.and(mask)
	a1 := fromUint128(first128, af)
	first = wrap(a1, &buf)

	// Calculate last IP in range: first | ^mask
	last128 := first128.or(mask.not())
	a2 := fromUint128(last128, af)
	last = wrap(a2, &buf)

	return
}

// Prefix tries to determine if the inclusive range [first, last]
// can be exactly represented as a single netip.Prefix.
// It returns the prefix and ok=true if so.
//
// Returns ok=false for ranges that don't align exactly to a prefix,
// invalid IPs, mismatched IP versions or first > last.
func Prefix(first, last netip.Addr) (prefix netip.Prefix, ok bool) {
	// Validity checks are performed via value-receiver methods first while
	// the address values can still be optimized in CPU registers.
	if !first.IsValid() || !last.IsValid() {
		return
	}

	// Check address family consistency, 4in6 is handled as v6.
	if first.Is4() != last.Is4() {
		return
	}

	// Extract internal representation of the prefix address as an addr struct
	// only after value-based inspection is complete.
	a := unwrap(first) // Low-level uint128 view of first address
	b := unwrap(last)  // Low-level uint128 view of last address

	// Ensure ordering: first <= last
	if a.ip.compare(b.ip) == 1 {
		return
	}

	// Determine prefix length and validity for exact CIDR match
	bits, ok := a.ip.prefixOK(b.ip)
	if !ok {
		return
	}

	if first.Is4() {
		// For IPv4 mapped in IPv6 space, adjust prefix length
		bits -= 96
	}

	// Construct prefix from first IP and prefix length bits.
	return netip.PrefixFrom(first, bits), ok
}

// All returns an iterator over all netip.Prefix values that
// cover the entire inclusive IP range [first, last].
//
// If either IP is invalid, the order is wrong, or IP versions differ,
// the iterator yields no results.
//
// This uses a recursive subdivision approach to partition
// the range into a minimal set of CIDRs.
func All(first, last netip.Addr) iter.Seq[netip.Prefix] {
	// invalid IP
	if !first.IsValid() || !last.IsValid() {
		return zeroIter
	}

	// Check address family consistency, 4in6 is handled as v6.
	if first.Is4() != last.Is4() {
		return zeroIter
	}
	a := unwrap(first) // low-level uint128 view of first
	b := unwrap(last)  // low-level uint128 view of last

	// Ensure ordering: first <= last
	if a.ip.compare(b.ip) == 1 {
		return zeroIter
	}

	af := v4
	if !first.Is4() {
		af = v6
	}

	// happy path
	return func(yield func(netip.Prefix) bool) {
		// Start recursive subdivision and yield prefixes
		allRec(af, a, b, yield)
	}
}

// allRec recursively yields prefixes for the IP range [a, b].
//
// If the range [a, b] exactly matches a prefix, yields that prefix.
//
// Otherwise recursively splits the range into halves and processes both.
//
// All bit arithmetic and masking is done in uint128 space.
func allRec(af addrFamily, a, b addr, yield func(netip.Prefix) bool) bool {
	// Check if [a, b] is exactly a prefix range
	lcp, ok := a.ip.prefixOK(b.ip)
	if ok {
		// Found exact CIDR match - yield it and stop recursion
		if af == v4 {
			lcp -= 96 // Adjust for IPv4-in-IPv6 embedding
		}
		return yield(netip.PrefixFrom(wrap(a, &[16]byte{}), lcp))
	}

	// Range doesn't match a single CIDR - split it in half
	// lcp < 128 since prefixOK returned ok=false (a != b), ensuring lcp+1 <= 128 for mask6 indexing
	mask := mask6[lcp+1]                              // Mask for one bit longer prefix
	leftUpper := fromUint128(a.ip.or(mask.not()), af) // Left half upper bound
	rightLower := fromUint128(b.ip.and(mask), af)     // Right half lower bound

	// Recursively process both halves
	return allRec(af, a, leftUpper, yield) && allRec(af, rightLower, b, yield)
}

// CommonPrefix returns the smallest common supernet (longest common prefix)
// that covers both pfx1 and pfx2 entirely.
//
// It returns the zero value if either prefix is invalid or if the IP
// versions do not match.
func CommonPrefix(pfx1, pfx2 netip.Prefix) (pfx netip.Prefix) {
	if !pfx1.IsValid() || !pfx2.IsValid() {
		return
	}

	ip1 := pfx1.Addr()
	ip2 := pfx2.Addr()

	bits1 := pfx1.Bits()
	bits2 := pfx2.Bits()

	ip1Is4 := ip1.Is4()
	ip2Is4 := ip2.Is4()

	// Check address family consistency, 4in6 is handled as v6.
	if ip1Is4 != ip2Is4 {
		return
	}

	addr1 := unwrap(ip1)
	addr2 := unwrap(ip2)

	// count matching bits in the full 128-bit space
	lcp := addr1.ip.commonPrefixLen(addr2.ip)
	if ip1Is4 {
		lcp -= 96 // adjust offset for IPv4
	}

	// final length must not exceed shortest input prefix
	minBits := min(lcp, bits1, bits2)

	// return normalized prefix; addr1 or addr2 works equally
	return netip.PrefixFrom(ip1, minBits).Masked()
}
