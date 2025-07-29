// Package extnetip is an extension to net/netip providing
// auxiliary functions for converting IP prefixes to IP ranges
// and vice versa.
//
// The calculations are done efficiently in uint128 space,
// avoiding conversions to/from byte slices.
//
// These extensions allow easy implementation of third-party libraries
// for IP range management on top of net/netip.
package extnetip

import (
	"iter"
	"net/netip"
)

// Range returns the inclusive IP address range [first, last]
// covered by the given prefix p.
//
// The prefix p does not have to be canonical.
//
// If p is invalid, Range returns zero values.
//
// The range calculation is performed by masking the uint128
// representation according to the prefix bits.
func Range(p netip.Prefix) (first, last netip.Addr) {
	if !p.IsValid() {
		return
	}

	// Extract internal representation of the prefix address as addr struct
	pa := unwrap(p.Addr())

	bits := p.Bits()
	if pa.is4() {
		// IPv4 addresses are embedded in IPv6 space with a 96-bit prefix
		bits += 96
	}
	mask := mask6(bits) // get the network mask as uint128

	// Calculate first IP in range: ip & mask
	first128 := pa.ip.and(mask)

	// Calculate last IP in range: first | ^mask
	last128 := first128.or(mask.not())

	// wrap back to netip.Addr, preserving IPv4 or IPv6 form
	first = wrap(fromUint128(first128, pa.is4()))
	last = wrap(fromUint128(last128, pa.is4()))

	return
}

// Prefix tries to determine if the inclusive range [first, last]
// can be exactly represented as a single netip.Prefix.
// It returns the prefix and ok=true if so.
//
// Returns ok=false for ranges that don't align exactly to a prefix,
// invalid IPs, mismatched versions or first > last.
//
// The calculation is done by analyzing the uint128 values
// and checking prefix match conditions.
func Prefix(first, last netip.Addr) (prefix netip.Prefix, ok bool) {
	// invalid IP
	if !first.IsValid() || !last.IsValid() {
		return
	}

	a := unwrap(first) // low-level uint128 view of first
	b := unwrap(last)  // low-level uint128 view of last

	// Check address family consistency.
	if a.is4() != b.is4() {
		return
	}

	// Ensure ordering: first <= last
	if a.ip.compare(b.ip) == 1 {
		return
	}

	// Determine prefix length and validity for exact CIDR match
	bits, ok := a.ip.prefixOK(b.ip)
	if !ok {
		return
	}

	if a.is4() {
		// For IPv4 mapped in IPv6 space, adjust prefix length
		bits -= 96
	}

	// Construct prefix from first IP and prefix length bits.
	return netip.PrefixFrom(first, bits), ok
}

// All returns an iterator over all netip.Prefix values that
// cover the entire inclusive IP range [first, last].
//
// If either IP is invalid, the order is wrong, or versions differ,
// the iterator yields no results.
//
// This uses a recursive subdivision approach to partition
// the range into a minimal set of CIDRs.
func All(first, last netip.Addr) iter.Seq[netip.Prefix] {
	return func(yield func(netip.Prefix) bool) {
		// invalid IP
		if !first.IsValid() || !last.IsValid() {
			return
		}

		a := unwrap(first) // low-level uint128 view of first
		b := unwrap(last)  // low-level uint128 view of last

		// Check address family consistency.
		if a.is4() != b.is4() {
			return
		}

		// Ensure ordering: first <= last
		if a.ip.compare(b.ip) == 1 {
			return
		}

		// Start recursive subdivision and yield prefixes
		allRec(a, b, yield)
	}
}

// allRec recursively yields prefixes for the IP range [a, b].
//
// If the range [a, b] exactly matches a prefix, yields that prefix.
//
// Otherwise recursively splits the range into halves and processes both.
//
// All bit arithmetic and masking is done in uint128 space.
func allRec(a, b addr, yield func(netip.Prefix) bool) bool {
	// Check if [a, b] is exactly a prefix range
	bits, ok := a.ip.prefixOK(b.ip)
	if ok {
		// recursion stop condition:
		if a.is4() {
			bits -= 96
		}
		return yield(netip.PrefixFrom(wrap(a), bits))
	}

	// If not an exact prefix, split the range for further subdivision

	// Calculate mask for one bit longer prefix length (to create a split point)
	mask := mask6(bits + 1)

	// Create midpoint by setting host bits for left half upper bound
	m1 := fromUint128(a.ip.or(mask.not()), a.is4())

	// Create midpoint by clearing host bits for right half lower bound
	m2 := fromUint128(b.ip.and(mask), a.is4())

	// Recursively yield prefixes for left half and then right half
	return allRec(a, m1, yield) && allRec(m2, b, yield)
}

// Deprecated: Prefixes is deprecated. Use the iterator version [All] instead.
func Prefixes(first, last netip.Addr) []netip.Prefix {
	return PrefixesAppend(nil, first, last)
}

// Deprecated: PrefixesAppend is deprecated. Use the iterator version [All] instead.
func PrefixesAppend(dst []netip.Prefix, first, last netip.Addr) []netip.Prefix {
	for pfx := range All(first, last) {
		dst = append(dst, pfx)
	}
	return dst
}
