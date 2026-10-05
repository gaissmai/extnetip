package extnetip

import "math/bits"

// uint128 represents a 128-bit unsigned integer value using two uint64 parts.
//
// This struct models the internal numeric representation of netip.Addr
// in Go 1.18 and newer, where IP addresses are handled as 128-bit values.
type uint128 struct {
	hi uint64
	lo uint64
}

// and returns the bitwise AND of two uint128 values.
func (u uint128) and(v uint128) uint128 {
	return uint128{u.hi & v.hi, u.lo & v.lo}
}

// or returns the bitwise OR of two uint128 values.
func (u uint128) or(v uint128) uint128 {
	return uint128{u.hi | v.hi, u.lo | v.lo}
}

// xor returns the bitwise XOR of two uint128 values.
func (u uint128) xor(v uint128) uint128 {
	return uint128{u.hi ^ v.hi, u.lo ^ v.lo}
}

// not returns the bitwise complement (inversion) of the uint128 value.
func (u uint128) not() uint128 {
	return uint128{^u.hi, ^u.lo}
}

// commonPrefixLen returns the number of leading bits that two uint128
// values have in common.
func (u uint128) commonPrefixLen(v uint128) int {
	commonHi := bits.LeadingZeros64(u.hi ^ v.hi)
	commonLo := bits.LeadingZeros64(u.lo ^ v.lo)

	if commonHi != 64 {
		return commonHi
	}
	return commonLo + 64
}

// prefixOK checks if the range from u to v (inclusive) forms an exact IP prefix (CIDR block).
//
// Returns:
//   - lcp: the length of the common prefix bits between u and v.
//   - ok: true if [u, v] corresponds exactly to a CIDR range without gaps.
//
// The check confirms:
//   - That the prefix length matches the differing bits.
//   - That u has zeros in all host bits (the prefix network part).
//   - That v has ones in all host bits (the prefix broadcast/end address).
func (u uint128) prefixOK(v uint128) (lcp int, ok bool) {
	lcp = u.commonPrefixLen(v)
	if lcp == 128 {
		return lcp, true
	}
	mask := mask6[lcp]

	// check if mask applied to first and last results in all zeros and all ones
	allZero := u.xor(u.and(mask)) == uint128{}
	allOnes := v.or(mask) == uint128{^uint64(0), ^uint64(0)}

	return lcp, allZero && allOnes
}

// compare compares u and v and returns:
//
//	 1 if u > v
//	-1 if u < v
//	 0 if u == v
func (u uint128) compare(v uint128) int {
	if u.hi > v.hi {
		return 1
	}
	if u.hi < v.hi {
		return -1
	}

	// u.hi == v.hi
	if u.lo > v.lo {
		return 1
	}
	if u.lo < v.lo {
		return -1
	}

	// u.hi == v.hi && u.lo == v.lo
	return 0
}

// isZero reports whether u represents the value zero.
func (u uint128) isZero() bool {
	return u.hi|u.lo == 0
}

// mask6 is a precalculated to keep Contains6 inlined
var mask6 = [129]uint128{
	0:   {0x0000000000000000, 0x0000000000000000},
	1:   {0x8000000000000000, 0x0000000000000000},
	2:   {0xc000000000000000, 0x0000000000000000},
	3:   {0xe000000000000000, 0x0000000000000000},
	4:   {0xf000000000000000, 0x0000000000000000},
	5:   {0xf800000000000000, 0x0000000000000000},
	6:   {0xfc00000000000000, 0x0000000000000000},
	7:   {0xfe00000000000000, 0x0000000000000000},
	8:   {0xff00000000000000, 0x0000000000000000},
	9:   {0xff80000000000000, 0x0000000000000000},
	10:  {0xffc0000000000000, 0x0000000000000000},
	11:  {0xffe0000000000000, 0x0000000000000000},
	12:  {0xfff0000000000000, 0x0000000000000000},
	13:  {0xfff8000000000000, 0x0000000000000000},
	14:  {0xfffc000000000000, 0x0000000000000000},
	15:  {0xfffe000000000000, 0x0000000000000000},
	16:  {0xffff000000000000, 0x0000000000000000},
	17:  {0xffff800000000000, 0x0000000000000000},
	18:  {0xffffc00000000000, 0x0000000000000000},
	19:  {0xffffe00000000000, 0x0000000000000000},
	20:  {0xfffff00000000000, 0x0000000000000000},
	21:  {0xfffff80000000000, 0x0000000000000000},
	22:  {0xfffffc0000000000, 0x0000000000000000},
	23:  {0xfffffe0000000000, 0x0000000000000000},
	24:  {0xffffff0000000000, 0x0000000000000000},
	25:  {0xffffff8000000000, 0x0000000000000000},
	26:  {0xffffffc000000000, 0x0000000000000000},
	27:  {0xffffffe000000000, 0x0000000000000000},
	28:  {0xfffffff000000000, 0x0000000000000000},
	29:  {0xfffffff800000000, 0x0000000000000000},
	30:  {0xfffffffc00000000, 0x0000000000000000},
	31:  {0xfffffffe00000000, 0x0000000000000000},
	32:  {0xffffffff00000000, 0x0000000000000000},
	33:  {0xffffffff80000000, 0x0000000000000000},
	34:  {0xffffffffc0000000, 0x0000000000000000},
	35:  {0xffffffffe0000000, 0x0000000000000000},
	36:  {0xfffffffff0000000, 0x0000000000000000},
	37:  {0xfffffffff8000000, 0x0000000000000000},
	38:  {0xfffffffffc000000, 0x0000000000000000},
	39:  {0xfffffffffe000000, 0x0000000000000000},
	40:  {0xffffffffff000000, 0x0000000000000000},
	41:  {0xffffffffff800000, 0x0000000000000000},
	42:  {0xffffffffffc00000, 0x0000000000000000},
	43:  {0xffffffffffe00000, 0x0000000000000000},
	44:  {0xfffffffffff00000, 0x0000000000000000},
	45:  {0xfffffffffff80000, 0x0000000000000000},
	46:  {0xfffffffffffc0000, 0x0000000000000000},
	47:  {0xfffffffffffe0000, 0x0000000000000000},
	48:  {0xffffffffffff0000, 0x0000000000000000},
	49:  {0xffffffffffff8000, 0x0000000000000000},
	50:  {0xffffffffffffc000, 0x0000000000000000},
	51:  {0xffffffffffffe000, 0x0000000000000000},
	52:  {0xfffffffffffff000, 0x0000000000000000},
	53:  {0xfffffffffffff800, 0x0000000000000000},
	54:  {0xfffffffffffffc00, 0x0000000000000000},
	55:  {0xfffffffffffffe00, 0x0000000000000000},
	56:  {0xffffffffffffff00, 0x0000000000000000},
	57:  {0xffffffffffffff80, 0x0000000000000000},
	58:  {0xffffffffffffffc0, 0x0000000000000000},
	59:  {0xffffffffffffffe0, 0x0000000000000000},
	60:  {0xfffffffffffffff0, 0x0000000000000000},
	61:  {0xfffffffffffffff8, 0x0000000000000000},
	62:  {0xfffffffffffffffc, 0x0000000000000000},
	63:  {0xfffffffffffffffe, 0x0000000000000000},
	64:  {0xffffffffffffffff, 0x0000000000000000},
	65:  {0xffffffffffffffff, 0x8000000000000000},
	66:  {0xffffffffffffffff, 0xc000000000000000},
	67:  {0xffffffffffffffff, 0xe000000000000000},
	68:  {0xffffffffffffffff, 0xf000000000000000},
	69:  {0xffffffffffffffff, 0xf800000000000000},
	70:  {0xffffffffffffffff, 0xfc00000000000000},
	71:  {0xffffffffffffffff, 0xfe00000000000000},
	72:  {0xffffffffffffffff, 0xff00000000000000},
	73:  {0xffffffffffffffff, 0xff80000000000000},
	74:  {0xffffffffffffffff, 0xffc0000000000000},
	75:  {0xffffffffffffffff, 0xffe0000000000000},
	76:  {0xffffffffffffffff, 0xfff0000000000000},
	77:  {0xffffffffffffffff, 0xfff8000000000000},
	78:  {0xffffffffffffffff, 0xfffc000000000000},
	79:  {0xffffffffffffffff, 0xfffe000000000000},
	80:  {0xffffffffffffffff, 0xffff000000000000},
	81:  {0xffffffffffffffff, 0xffff800000000000},
	82:  {0xffffffffffffffff, 0xffffc00000000000},
	83:  {0xffffffffffffffff, 0xffffe00000000000},
	84:  {0xffffffffffffffff, 0xfffff00000000000},
	85:  {0xffffffffffffffff, 0xfffff80000000000},
	86:  {0xffffffffffffffff, 0xfffffc0000000000},
	87:  {0xffffffffffffffff, 0xfffffe0000000000},
	88:  {0xffffffffffffffff, 0xffffff0000000000},
	89:  {0xffffffffffffffff, 0xffffff8000000000},
	90:  {0xffffffffffffffff, 0xffffffc000000000},
	91:  {0xffffffffffffffff, 0xffffffe000000000},
	92:  {0xffffffffffffffff, 0xfffffff000000000},
	93:  {0xffffffffffffffff, 0xfffffff800000000},
	94:  {0xffffffffffffffff, 0xfffffffc00000000},
	95:  {0xffffffffffffffff, 0xfffffffe00000000},
	96:  {0xffffffffffffffff, 0xffffffff00000000},
	97:  {0xffffffffffffffff, 0xffffffff80000000},
	98:  {0xffffffffffffffff, 0xffffffffc0000000},
	99:  {0xffffffffffffffff, 0xffffffffe0000000},
	100: {0xffffffffffffffff, 0xfffffffff0000000},
	101: {0xffffffffffffffff, 0xfffffffff8000000},
	102: {0xffffffffffffffff, 0xfffffffffc000000},
	103: {0xffffffffffffffff, 0xfffffffffe000000},
	104: {0xffffffffffffffff, 0xffffffffff000000},
	105: {0xffffffffffffffff, 0xffffffffff800000},
	106: {0xffffffffffffffff, 0xffffffffffc00000},
	107: {0xffffffffffffffff, 0xffffffffffe00000},
	108: {0xffffffffffffffff, 0xfffffffffff00000},
	109: {0xffffffffffffffff, 0xfffffffffff80000},
	110: {0xffffffffffffffff, 0xfffffffffffc0000},
	111: {0xffffffffffffffff, 0xfffffffffffe0000},
	112: {0xffffffffffffffff, 0xffffffffffff0000},
	113: {0xffffffffffffffff, 0xffffffffffff8000},
	114: {0xffffffffffffffff, 0xffffffffffffc000},
	115: {0xffffffffffffffff, 0xffffffffffffe000},
	116: {0xffffffffffffffff, 0xfffffffffffff000},
	117: {0xffffffffffffffff, 0xfffffffffffff800},
	118: {0xffffffffffffffff, 0xfffffffffffffc00},
	119: {0xffffffffffffffff, 0xfffffffffffffe00},
	120: {0xffffffffffffffff, 0xffffffffffffff00},
	121: {0xffffffffffffffff, 0xffffffffffffff80},
	122: {0xffffffffffffffff, 0xffffffffffffffc0},
	123: {0xffffffffffffffff, 0xffffffffffffffe0},
	124: {0xffffffffffffffff, 0xfffffffffffffff0},
	125: {0xffffffffffffffff, 0xfffffffffffffff8},
	126: {0xffffffffffffffff, 0xfffffffffffffffc},
	127: {0xffffffffffffffff, 0xfffffffffffffffe},
	128: {0xffffffffffffffff, 0xffffffffffffffff},
}
