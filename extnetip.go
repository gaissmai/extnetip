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

type addrFamily uint

const (
	invalid addrFamily = iota
	v4
	v6
	// v4in6 handled as v6
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
	is4 := ip.Is4()

	// Extract internal representation of the prefix address as addr struct
	pa := unwrap(ip)

	bits := p.Bits()
	if is4 {
		// IPv4 addresses are embedded in IPv6 space with a 96-bit prefix
		bits += 96
	}
	mask := mask6[bits] // get the network mask as uint128

	// Calculate first IP in range: ip & mask
	first128 := pa.ip.and(mask)

	// Calculate last IP in range: first | ^mask
	last128 := first128.or(mask.not())

	// wrap back to netip.Addr, preserving IPv4 or IPv6 form
	first = wrap(fromUint128(first128, is4), &[16]byte{})
	last = wrap(fromUint128(last128, is4), &[16]byte{})

	return
}

// Prefix tries to determine if the inclusive range [first, last]
// can be exactly represented as a single netip.Prefix.
// It returns the prefix and ok=true if so.
//
// Returns ok=false for ranges that don't align exactly to a prefix,
// invalid IPs, mismatched IP versions or first > last.
func Prefix(first, last netip.Addr) (prefix netip.Prefix, ok bool) {
	// check validity
	if !first.IsValid() || !last.IsValid() {
		return
	}

	// Check address family consistency, 4in6 is handled as v6.
	if first.Is4() != last.Is4() {
		return
	}

	a := unwrap(first) // low-level uint128 view of first
	b := unwrap(last)  // low-level uint128 view of last

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

	// happy path
	return func(yield func(netip.Prefix) bool) {
		// Start recursive subdivision and yield prefixes
		allRec(first.Is4(), a, b, yield)
	}
}

// allRec recursively yields prefixes for the IP range [a, b].
//
// If the range [a, b] exactly matches a prefix, yields that prefix.
//
// Otherwise recursively splits the range into halves and processes both.
//
// All bit arithmetic and masking is done in uint128 space.
func allRec(is4 bool, a, b addrT, yield func(netip.Prefix) bool) bool {
	// Check if [a, b] is exactly a prefix range
	lcp, ok := a.ip.prefixOK(b.ip)
	if ok {
		// Found exact CIDR match - yield it and stop recursion
		if is4 {
			lcp -= 96 // Adjust for IPv4-in-IPv6 embedding
		}
		return yield(netip.PrefixFrom(wrap(a, &[16]byte{}), lcp))
	}

	// Range doesn't match a single CIDR - split it in half
	// lcp < 128 since prefixOK returned ok=false (a != b), ensuring lcp+1 <= 128 for mask6 indexing
	mask := mask6[lcp+1]                               // Mask for one bit longer prefix
	leftUpper := fromUint128(a.ip.or(mask.not()), is4) // Left half upper bound
	rightLower := fromUint128(b.ip.and(mask), is4)     // Right half lower bound

	// Recursively process both halves
	return allRec(is4, a, leftUpper, yield) && allRec(is4, rightLower, b, yield)
}

// CommonPrefix returns the longest prefix shared by pfx1 and pfx2.
// It returns the zero value if a prefix is invalid or if the IP
// versions do not match. Otherwise it compares both addresses
// and returns the prefix covering their common range.
func CommonPrefix(pfx1, pfx2 netip.Prefix) (pfx netip.Prefix) {
	if !pfx1.IsValid() || !pfx2.IsValid() {
		return
	}

	addr1 := pfx1.Masked().Addr()
	addr2 := pfx2.Masked().Addr()

	is4 := addr1.Is4()
	if addr2.Is4() != is4 {
		return
	}

	ext1 := unwrap(addr1)
	ext2 := unwrap(addr2)

	// count matching bits in the full 128-bit space
	commonBits := ext1.ip.commonPrefixLen(ext2.ip)
	if is4 {
		commonBits -= 96 // adjust offset for IPv4
	}

	// final length must not exceed shorter input prefix
	minBits := min(commonBits, pfx1.Bits(), pfx2.Bits())

	// return normalized prefix; addr1 or addr2 works equally
	return netip.PrefixFrom(addr1, minBits).Masked()
}
