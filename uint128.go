package extnetip

import "math/bits"

// the internal representation of netip.Addr is mainly an uint128 (at go version 1.18)

type uint128 struct {
	hi uint64
	lo uint64
}

func (u uint128) and(v uint128) uint128 {
	return uint128{u.hi & v.hi, u.lo & v.lo}
}

func (u uint128) or(v uint128) uint128 {
	return uint128{u.hi | v.hi, u.lo | v.lo}
}

func (u uint128) xor(v uint128) uint128 {
	return uint128{u.hi ^ v.hi, u.lo ^ v.lo}
}

func (u uint128) not() uint128 {
	return uint128{^u.hi, ^u.lo}
}

func mask128(n int) uint128 {
	return uint128{^(^uint64(0) >> n), ^uint64(0) << (128 - n)}
}

func u64PrefixLenInCommon(a, b uint64) int {
	return bits.LeadingZeros64(a ^ b)
}

func (u uint128) prefixLenInCommon(v uint128) (n int) {
	if n = u64PrefixLenInCommon(u.hi, v.hi); n == 64 {
		n += u64PrefixLenInCommon(u.lo, v.lo)
	}
	return
}

// prefixOK returns the common bits of two uint128 and true if they present exactly a prefix.
func (a uint128) prefixOK(b uint128) (bits int, ok bool) {
	bits = a.prefixLenInCommon(b)
	if bits == 128 {
		return bits, true
	}
	mask := mask128(bits)

	// check if mask applied to base and last results in all zeros and all ones
	allZero := a.xor(a.and(mask)) == uint128{}
	allOnes := b.or(mask) == uint128{^uint64(0), ^uint64(0)}

	return bits, allZero && allOnes
}
