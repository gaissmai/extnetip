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

// addrT holds the IP address as uint128 data and a flag if it is IPv4.
//
// This struct is used for arithmetic or comparison operations
// on netip.Addr data in a safe manner.
type addrT struct {
	ip uint128
	af addrFamily
}

// fromUint128 creates an addr struct from a uint128 IP representation and a flag is4.
func fromUint128(ip uint128, is4 bool) addrT {
	if is4 {
		return addrT{ip: ip, af: v4}
	}
	return addrT{ip: ip, af: v6}
}

// unwrap extracts the raw uint128 representation from a netip.Addr safely.
func unwrap(a netip.Addr) (b addrT) {
	octets := a.AsSlice()
	if octets == nil {
		return
	}

	if len(octets) == 4 {
		b.af = v4
		b.ip.lo = uint64(binary.BigEndian.Uint32(octets))
		return b
	}

	b.af = v6
	b.ip.hi = binary.BigEndian.Uint64(octets[:8])
	b.ip.lo = binary.BigEndian.Uint64(octets[8:])

	return b
}

// wrap converts an addr back to a netip.Addr using safe conversions via byte arrays.
func wrap(a addrT, buf *[16]byte) (ip netip.Addr) {
	if a.af == invalid {
		return
	}

	binary.BigEndian.PutUint64(buf[8:], a.ip.lo)
	if a.af == v4 {
		return netip.AddrFrom4([4]byte(buf[12:]))
	}

	binary.BigEndian.PutUint64(buf[:8], a.ip.hi)
	return netip.AddrFrom16(*buf)
}
