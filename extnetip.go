// Package extnetip is an extension to net/netip with
// a few missing but important auxiliary functions for
// converting IP-prefixes to IP-ranges and vice versa.
//
// The functions are effectively performed in uint128 space,
// no conversions from/to bytes are performed.
//
// With these extensions to net/netip, third-party IP-range
// libraries become easily possible.
package extnetip

import (
	"iter"
	"net/netip"
)

// Range returns the inclusive range of IP addresses [first, last] that p covers.
//
// If p is invalid, Range returns the zero values.
func Range(p netip.Prefix) (first, last netip.Addr) {
	if !p.IsValid() {
		return
	}

	// peek the internals, do math in uint128
	pa := peek(p.Addr())

	bits := p.Bits()
	if pa.is4 {
		bits += 96
	}
	mask := mask6(bits)

	first128 := pa.ip.and(mask)
	last128 := first128.or(mask.not())

	// convert back to netip.Addr
	first = back(addr{first128, pa.is4})
	last = back(addr{last128, pa.is4})

	return
}

// Prefix returns the netip.Prefix from first to last and ok=true,
// if it can be presented exactly as such.
//
// If first or last are not valid, in the wrong order or not exactly
// equal to one prefix, ok is false.
func Prefix(first, last netip.Addr) (prefix netip.Prefix, ok bool) {
	// wrong input
	switch {
	case !(first.IsValid() && last.IsValid()): // invalid IP
		return
	case first.Is4() != last.Is4(): // different version
		return
	case last.Less(first): // wrong order
		return
	}

	// peek the internals, do math in uint128
	a := peek(first)
	b := peek(last)

	// IP versions differ?
	if a.is4 != b.is4 {
		return
	}

	// do math in uint128
	bits, ok := a.ip.prefixOK(b.ip)
	if !ok {
		return
	}

	if a.is4 {
		bits -= 96
	}

	// make prefix, possible zone gets dropped
	return netip.PrefixFrom(first, bits), ok
}

// All returns an iterator over the set of prefixes that covers the range from [first, last].
//
// If first or last are not valid, in the wrong order or not of the same version, the set is empty.
func All(first, last netip.Addr) iter.Seq[netip.Prefix] {
	return func(yield func(netip.Prefix) bool) {
		// wrong input
		switch {
		case !(first.IsValid() && last.IsValid()): // invalid IP
			return
		case first.Is4() != last.Is4(): // different version
			return
		case last.Less(first): // wrong order
			return
		}

		allRec(peek(first), peek(last), yield)
	}
}

// allRec yields the prefix if [a, b] represents a whole CIDR, like 10.0.0.0/8
// (first being 10.0.0.0 and last being 10.255.255.255)
//
// Otherwise recursively do both halves. All bit fiddling calculations are done in the uint128 space.
func allRec(a, b addr, yield func(netip.Prefix) bool) bool {
	// recursion stop condition, [a,b] already representing a prefix
	bits, ok := a.ip.prefixOK(b.ip)
	if ok {
		if a.is4 {
			bits -= 96
		}
		return yield(netip.PrefixFrom(back(a), bits))
	}

	// otherwise split the range, make two halves and do both halves recursively
	mask := mask6(bits + 1)

	// set hostbits, make middle left
	m1 := addr{a.ip.or(mask.not()), a.is4}

	// clear hostbits, make middle right
	m2 := addr{b.ip.and(mask), a.is4}

	// ... do both halves recursively
	return allRec(a, m1, yield) && allRec(m2, b, yield)
}

// Prefixes returns the set of netip.Prefix entries that covers the
// IP range from first to last.
//
// If first or last are invalid, in the wrong order, or if they're of different
// address families, then Prefixes returns nil.
//
// Deprecated: Prefixes is deprecated. Use the iterator version [All] instead.
func Prefixes(first, last netip.Addr) []netip.Prefix {
	return PrefixesAppend(nil, first, last)
}

// PrefixesAppend is an append version of Prefixes. It appends
// the netip.Prefix entries to dst that covers the IP range from first to last.
//
// Deprecated: PrefixesAppend is deprecated. Use the iterator version [All] instead.
func PrefixesAppend(dst []netip.Prefix, first, last netip.Addr) []netip.Prefix {
	for pfx := range All(first, last) {
		dst = append(dst, pfx)
	}
	return dst
}
