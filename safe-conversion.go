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
	ip    uint128
	v4    bool
	valid bool
}

// fromUint128 creates an addr struct from a uint128 IP representation and a flag is4.
func fromUint128(ip uint128, is4 bool) addrT {
	return addrT{ip: ip, v4: is4, valid: true}
}

// is4 returns true if the address represents an IPv4 value.
func (a *addrT) is4() bool {
	return a.valid && a.v4
}

// unwrap extracts the raw uint128 representation from a netip.Addr safely.
func unwrap(a netip.Addr) (b addrT) {
	if !a.IsValid() {
		return b
	}
	b.valid = true

	ip16 := a.As16()

	b.ip.lo = binary.BigEndian.Uint64(ip16[8:])

	if a.Is4() {
		b.v4 = true
		return
	}

	b.ip.hi = binary.BigEndian.Uint64(ip16[:8])

	return b
}

// wrap converts an addr back to a netip.Addr using safe conversions via byte arrays.
func wrap(a addrT) netip.Addr {
	if !a.valid {
		return netip.Addr{}
	}

	var a16 [16]byte
	binary.BigEndian.PutUint64(a16[8:], a.ip.lo)

	if a.v4 {
		return netip.AddrFrom4([4]byte(a16[12:]))
	}

	binary.BigEndian.PutUint64(a16[:8], a.ip.hi)
	return netip.AddrFrom16(a16)
}
