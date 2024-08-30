//go:build go1.23

package extnetip

import "net/netip"

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
