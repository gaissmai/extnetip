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

func peek(a netip.Addr) (b addr) {
	b.is4 = a.Is4()

	raw := a.As16()
	b.ip.hi = binary.BigEndian.Uint64(raw[:8])
	b.ip.lo = binary.BigEndian.Uint64(raw[8:])
	return
}

func back(a addr) netip.Addr {
	var a6 [16]byte
	binary.BigEndian.PutUint64(a6[:8], a.ip.hi)
	binary.BigEndian.PutUint64(a6[8:], a.ip.lo)

	if a.is4 {
		// convert slice to array pointer
		a4 := (*[4]byte)(a6[12:])
		return netip.AddrFrom4(*a4)
	}

	return netip.AddrFrom16(a6)
}
