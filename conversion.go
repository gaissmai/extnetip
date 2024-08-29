package extnetip

import (
	"encoding/binary"
	"net/netip"
)

// addr is a struct for peeks into netip.Addr for uint128 math calculations.
type addr struct {
	ip  uint128
	is4 bool
}

func peek(a netip.Addr) addr {
	var b addr
	b.is4 = a.Is4()

	a16 := a.As16()
	b.ip.hi = binary.BigEndian.Uint64(a16[:8])
	b.ip.lo = binary.BigEndian.Uint64(a16[8:])
	return b
}

func back(a addr) netip.Addr {
	var a16 [16]byte
	binary.BigEndian.PutUint64(a16[:8], a.ip.hi)
	binary.BigEndian.PutUint64(a16[8:], a.ip.lo)

	if a.is4 {
		return netip.AddrFrom4([4]byte(a16[12:]))
	}

	return netip.AddrFrom16(a16)
}
