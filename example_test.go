//go:build go1.23

package extnetip_test

import (
	"fmt"
	"net/netip"

	"github.com/gaissmai/extnetip"
)

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
