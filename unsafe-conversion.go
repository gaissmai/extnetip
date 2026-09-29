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

// addrT is a low-level struct used to interpret netip.Addr internals via unsafe.Pointer.
//
// It contains:
// - ip: the raw 128-bit IP address data as a uint128.
// - z:  used internally by netip.Addr to track address kind and zone.
//
// This struct layout must match netip.Addr exactly for unsafe conversions to work.
type addrT struct {
	ip uint128
	z  unsafe.Pointer
}

// Internal singleton pointers extracted from zero-value netip.Addr instances.
// These uintptr values correspond to internal discriminators used by netip.Addr
// to distinguish the kind of IP address representation.
//
// z4    - IPv4 address representation
// z6noz - IPv6 address representation without zone
var (
	z4    unsafe.Pointer
	z6noz unsafe.Pointer
)

// Compile-time and runtime sanity checks: fail fast if layout changes.
// If sizes differ this line triggers a compile-time error (array length mismatch).
var _ [1]struct{} = [1 - int(unsafe.Sizeof(addrT{})-unsafe.Sizeof(netip.Addr{}))]struct{}{}

func init() {
	if unsafe.Alignof(addrT{}) != unsafe.Alignof(netip.Addr{}) {
		panic("extnetip: netip.Addr alignment changed; rebuild without -tags=unsafe")
	}

	// Initialize discriminators only after alignment is verified.
	z4 = unwrap(netip.AddrFrom4([4]byte{})).z
	z6noz = unwrap(netip.AddrFrom16([16]byte{})).z
}

// is4 checks whether the internal address representation corresponds to an IPv4 address.
//
// It compares the addr's z discriminator with the known IPv4 singleton z4.
func (a *addrT) is4() bool {
	return a.z == z4
}

// fromUint128 constructs an addr struct from a uint128 IP representation and a flag is4.
//
// If is4 is true, the addr assumes the IPv4 internal discriminator (z4).
// Otherwise, it uses the IPv6 no-zone internal discriminator (z6noz).
func fromUint128(ip uint128, is4 bool) addrT {
	if is4 {
		return addrT{ip, z4}
	}
	return addrT{ip, z6noz}
}

// unwrap casts a netip.Addr value into the internal addr representation.
func unwrap(ip netip.Addr) addrT {
	return *(*addrT)(unsafe.Pointer(&ip))
}

// wrap casts from the internal addr representation back to netip.Addr.
func wrap(a addrT) netip.Addr {
	return *(*netip.Addr)(unsafe.Pointer(&a))
}
