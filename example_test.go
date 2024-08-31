package extnetip_test

import (
	"fmt"
	"net/netip"

	"github.com/gaissmai/extnetip"
)

func ExampleRange() {
	pfx := netip.MustParsePrefix("fe80::/10")
	first, last := extnetip.Range(pfx)

	fmt.Println("First:", first)
	fmt.Println("Last: ", last)
	// Output:
	// First: fe80::
	// Last:  febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff
}

func ExamplePrefix() {
	first := netip.MustParseAddr("fe80::")
	last := netip.MustParseAddr("fe80::7")

	pfx, ok := extnetip.Prefix(first, last)

	fmt.Println("OK:    ", ok)
	fmt.Println("Prefix:", pfx)

	fmt.Println()

	first = netip.MustParseAddr("10.0.0.1")
	last = netip.MustParseAddr("10.0.0.19")

	pfx, ok = extnetip.Prefix(first, last)

	fmt.Println("OK:    ", ok)
	fmt.Println("Prefix:", pfx)

	// Output:
	// OK:     true
	// Prefix: fe80::/125
	//
	// OK:     false
	// Prefix: invalid Prefix
}

func ExampleAll() {
	first := netip.MustParseAddr("10.1.0.0")
	last := netip.MustParseAddr("10.1.13.233")

	fmt.Println("Prefixes:")
	for pfx := range extnetip.All(first, last) {
		fmt.Println(pfx)
	}

	// Output:
	// Prefixes:
	// 10.1.0.0/21
	// 10.1.8.0/22
	// 10.1.12.0/24
	// 10.1.13.0/25
	// 10.1.13.128/26
	// 10.1.13.192/27
	// 10.1.13.224/29
	// 10.1.13.232/31
}
