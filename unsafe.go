package extnetip

import (
	"net/netip"
	"unsafe"
)

// exhibType is a struct for unsafe peeks into netip.Addr for uint128 math calculations.
type exhibType struct {
	addr uint128
	z    uintptr
}

// peek the singleton pointer for z4 from netip
var (
	z4 uintptr = peek(netip.AddrFrom4([4]byte{})).z
)

// peek into the private internals of netip.Addr with unsafe.Pointer
func peek(ip netip.Addr) exhibType {
	return *(*exhibType)(unsafe.Pointer(&ip))
}

// back conversion to netip.Addr
func back(ip exhibType) netip.Addr {
	return *(*netip.Addr)(unsafe.Pointer(&ip))
}
