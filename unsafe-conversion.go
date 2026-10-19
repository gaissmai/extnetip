//go:build unsafe

// Package extnetip unsafe conversion implementation.
//
// This file is used when built with the 'unsafe' build tag.
// It provides fast conversions using unsafe.Pointer operations but requires
// that the internal layout of netip.Addr remains stable.
//
// WARNING: This implementation depends on internal Go standard library
// implementation details and may break with future Go versions.

package extnetip

import (
	"net/netip"
	"unsafe"
)

// addr is a low-level struct used to interpret netip.Addr internals via unsafe.Pointer.
//
// It contains:
// - ip: the raw 128-bit IP address data as a uint128.
// - z:  used internally by netip.Addr to track address kind and zone.
//
// This struct layout must match netip.Addr exactly for unsafe conversions to work.
type addr struct {
	ip uint128
	z  unsafe.Pointer
}

// Compile-time assertion to verify memory layout compatibility.
const (
	sizeDiff  = int(unsafe.Sizeof(addr{})) - int(unsafe.Sizeof(netip.Addr{}))
	alignDiff = int(unsafe.Alignof(addr{})) - int(unsafe.Alignof(netip.Addr{}))
)

// Use a bitwise-OR (sizeDiff | alignDiff) instead of addition (+)
// to prevent potential error compensation (e.g., a positive offset error
// cancelling out a negative one). If either size or alignment differs,
// the resulting value is non-zero, creating a wrong array length
// which triggers an immediate compile-time error.
var _ [1]struct{} = [1 - (sizeDiff | alignDiff)]struct{}{}

// Internal sentinel pointers extracted from netip.Addr instances.
// These pointer values correspond to internal sentinels used by netip.Addr
// to distinguish the kind of IP address representation.
//
// z4    - IPv4 address representation
// z6noz - IPv6 address representation without zone
var (
	z4    = unwrap(netip.IPv4Unspecified()).z
	z6noz = unwrap(netip.IPv6Unspecified()).z
)

// fromUint128 constructs an addr struct from a uint128 IP representation and af.
func fromUint128(ip uint128, af addrFamily) addr {
	if af == v4 {
		return addr{ip, z4}
	}
	return addr{ip, z6noz}
}

// unwrap casts a netip.Addr value into the internal addr representation.
func unwrap(ip netip.Addr) addr {
	return *(*addr)(unsafe.Pointer(&ip))
}

// wrap casts from the internal addr representation back to netip.Addr using unsafe.Pointer.
// The scratch buffer parameter is ignored in the unsafe implementation but kept for API parity
// with the safe build.
func wrap(a addr, _ *[16]byte) netip.Addr {
	return *(*netip.Addr)(unsafe.Pointer(&a))
}
