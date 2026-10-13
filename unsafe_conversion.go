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
// - z:  a uintptr used internally by netip.Addr to track address kind/discriminator.
//
// This struct layout must match netip.Addr exactly for unsafe conversions to work.
type addrT struct {
	ip uint128
	z  unsafe.Pointer
}

// prefixT mirrors the memory layout of net/netip.Prefix.
type prefixT struct {
	addr        addrT
	bitsPlusOne uint8
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

// is4 checks whether the internal address representation corresponds to an IPv4 address.
//
// It compares the addr's z discriminator with the known IPv4 singleton z4.
func (a *addrT) is4() bool {
	return a.z == z4
}

// As2xUint64 returns the IP address in its 128-bit representation.
// IPv4 addresses are returned as IPv4-mapped IPv6 addresses.
// IPv6 addresses with zones are returned without their zone
// (use the [netip.Addr.Zone] method to get it).
// The ip zero value returns all zeroes.
func as2xUint64(ip netip.Addr) (hi, lo uint64) {
	addr := unwrap(ip)
	return addr.ip.hi, addr.ip.lo
}

// AddrFrom2xUint64 returns the IPv6 address given by the 2 uint64 words. An
// IPv4-mapped IPv6 address is left as an IPv6 address. (Use Unmap to convert
// them if needed.)
func addrFrom2xUint64(hi, lo uint64) netip.Addr {
	a := addrT{ip: uint128{hi: hi, lo: lo}, z: z6noz}
	return *(*netip.Addr)(unsafe.Pointer(&a))
}

// containsUnsafe reports whether the network p includes ip using unsafe
// low-level memory reinterpretation for maximum performance.
//
// Unlike netip.Prefix.Contains, this function differs in the following ways:
//   - No Validation: Performs NO validity checks on the prefix or IP address.
//     Passing invalid or zero-value inputs can lead to undefined behavior.
//   - Zone Handling: Does NOT reject IPv6 addresses with zone identifiers.
//     Because it compares raw 128-bit address fields directly, an IPv6 address
//     with a zone can still match, whereas standard netip.Prefix.Contains
//     explicitly returns false if an IP has a zone.
//
// Callers must ensure that both p and ip are valid and compatible prior to invocation.
func containsUnsafe(p netip.Prefix, ip netip.Addr) bool {
	addr := *(*addrT)(unsafe.Pointer(&ip))
	pfx := *(*prefixT)(unsafe.Pointer(&p))
	bits := pfx.bitsPlusOne - 1

	if addr.z == z4 {
		bits += 96
	}
	return addr.ip.xor(pfx.addr.ip).and(mask6[bits]).isZero()
}

// unwrap converts a netip.Addr value into the internal addr representation using unsafe.Pointer.
//
// This is effectively a cast that allows direct access to netip.Addr internals without copying.
// Must only be used if struct layouts are confirmed to match.
//
// Precondition: a is a valid IP address.
func unwrap(a netip.Addr) addrT {
	return *(*addrT)(unsafe.Pointer(&a))
}

// wrap converts from the internal addr representation back to netip.Addr.
//
// It reconstructs a valid netip.Addr value by pointer-casting the addr.
//
// Use with caution; meaningful only if addr and netip.Addr share memory layout.
func wrap(a addrT) netip.Addr {
	return *(*netip.Addr)(unsafe.Pointer(&a))
}
