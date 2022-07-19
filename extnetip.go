// extnetip is an extension to net/netip.
//
// Some missing math functions are added to the closed
// private internals of netip using unsafe.
//
// No further types are defined, only helper functions on
// existing net/netip types.
//
// With these tiny extensions, third party IP-Range libraries,
// based on the stdlib net/netip, are now possible without
// further bytes/bits fumbling.
package extnetip

import "net/netip"

// Range returns the inclusive range of IP addresses that p covers.
//
// If p is invalid, Range returns the zero values.
func Range(p netip.Prefix) (first, last netip.Addr) {
	if !p.IsValid() {
		return
	}

	// peek the internals, do math in uint128
	exhib := peek(p.Addr())
	z := exhib.z

	bits := p.Bits()
	if z == z4 {
		bits += 96
	}
	mask := mask128(bits)

	base128 := exhib.addr.and(mask)
	last128 := base128.or(mask.not())

	// convert back to netip.Addr
	first = back(exhibType{base128, z})
	last = back(exhibType{last128, z})

	return
}

// Prefix returns the netip.Prefix from first to last and ok=true,
// if it can be presented exactly as such.
//
// If first or last are not valid, in the wrong order or not exactly
// equal to one prefix, ok is false.
func Prefix(first, last netip.Addr) (prefix netip.Prefix, ok bool) {
	if !(first.IsValid() && last.IsValid()) {
		return
	}
	if last.Less(first) {
		return
	}

	// peek the internals, do math in uint128
	exhibBase := peek(first)
	exhibLast := peek(last)

	// IP versions differ?
	if exhibBase.z != exhibLast.z {
		return
	}

	bits, ok := exhibBase.addr.prefixOK(exhibLast.addr)
	if !ok {
		return
	}

	if exhibBase.z == z4 {
		bits -= 96
	}

	// convert back to netip
	return netip.PrefixFrom(back(exhibBase), bits), ok
}

// Prefixes returns the set of netip.Prefix entries that covers the
// IP range from first to last.
//
// If first or last are invalid, in the wrong order, or if they're of different
// address families, then Prefixes returns nil.
//
// Prefixes necessarily allocates. See AppendPrefixes for a version that
// uses memory you provide.
func Prefixes(first, last netip.Addr) []netip.Prefix {
	return AppendPrefixes(nil, first, last)
}

// AppendPrefixes is an append version of Prefixes. It appends
// the netip.Prefix entries to dst that covers the IP range from first to last.
func AppendPrefixes(dst []netip.Prefix, first, last netip.Addr) []netip.Prefix {
	if !(first.IsValid() && last.IsValid()) {
		return nil
	}
	if last.Less(first) {
		return nil
	}

	// peek the internals, do math in uint128
	exhibFirst := peek(first)
	exhibLast := peek(last)

	// different IP versions
	if exhibFirst.z != exhibLast.z {
		return nil
	}

	// no recursion, use an iterative algo with stack
	var stack []exhibType

	// push, params are the starting point
	stack = append(stack, exhibFirst, exhibLast)

	for len(stack) > 0 {

		// pop two addresses
		exhibLast := stack[len(stack)-1]
		exhibFirst := stack[len(stack)-2]
		stack = stack[:len(stack)-2]

		// are base-last already representing a prefix?
		bits, ok := exhibFirst.addr.prefixOK(exhibLast.addr)
		if ok {
			if exhibFirst.z == z4 {
				bits -= 96
			}
			// convert back to netip
			pfx := netip.PrefixFrom(back(exhibFirst), bits)

			dst = append(dst, pfx)
			continue
		}

		// Otherwise split the range, make two halves and push it on the stack
		mask := mask128(bits + 1)

		// make middle last, set hostbits
		exhibMidOne := exhibType{exhibFirst.addr.or(mask.not()), exhibFirst.z}

		// make middle base, clear hostbits
		exhibMidTwo := exhibType{exhibLast.addr.and(mask), exhibFirst.z}

		// push both halves (in reverse order, prefixes are then sorted)
		stack = append(stack, exhibMidTwo, exhibLast)
		stack = append(stack, exhibFirst, exhibMidOne)
	}

	return dst
}