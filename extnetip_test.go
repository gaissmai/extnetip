package extnetip_test

import (
	"fmt"
	"net/netip"
	"reflect"
	"testing"

	"github.com/gaissmai/extnetip"
)

var (
	mustAddr   = netip.MustParseAddr
	mustPrefix = netip.MustParsePrefix
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
	// Output:
	// OK:     true
	// Prefix: fe80::/125
}

func ExamplePrefixes() {
	first := netip.MustParseAddr("10.1.0.0")
	last := netip.MustParseAddr("10.1.13.233")
	pfxs := extnetip.Prefixes(first, last)

	fmt.Println("Prefixes:")
	for _, pfx := range pfxs {
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

func pfxSlice(pfxStrs ...string) (out []netip.Prefix) {
	for _, s := range pfxStrs {
		out = append(out, mustPrefix(s))
	}
	return
}

func TestRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in    netip.Prefix
		first netip.Addr
		last  netip.Addr
	}{
		{
			netip.Prefix{},
			netip.Addr{},
			netip.Addr{},
		},
		{
			mustPrefix("0.0.0.0/0"),
			mustAddr("0.0.0.0"),
			mustAddr("255.255.255.255"),
		},
		{
			mustPrefix("10.0.0.0/8"),
			mustAddr("10.0.0.0"),
			mustAddr("10.255.255.255"),
		},
		{
			mustPrefix("172.16.0.0/12"),
			mustAddr("172.16.0.0"),
			mustAddr("172.31.255.255"),
		},
		{
			mustPrefix("::ffff:0.0.0.0/96"),
			mustAddr("::ffff:0.0.0.0"),
			mustAddr("::ffff:255.255.255.255"),
		},
		{
			mustPrefix("::/0"),
			mustAddr("::"),
			mustAddr("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
		},
		{
			mustPrefix("fe80::/10"),
			mustAddr("fe80::"),
			mustAddr("febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
		},
	}

	for _, tt := range tests {
		first, last := extnetip.Range(tt.in)
		if first != tt.first {
			t.Fatalf("Range(%s), got first: %s, expected: %s", tt.in, first, tt.first)
		}
		if last != tt.last {
			t.Fatalf("Range(%s), got last: %s, expected: %s", tt.in, last, tt.last)
		}
	}
}

func TestPrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip1 netip.Addr
		ip2 netip.Addr
		p   netip.Prefix
		ok  bool
	}{
		{
			netip.Addr{},
			netip.Addr{},
			netip.Prefix{},
			false,
		},
		{
			mustAddr("0.0.0.0"), // wrong versions
			mustAddr("::"),
			netip.Prefix{},
			false,
		},
		{
			mustAddr("0.0.0.1"), // wrong order
			mustAddr("0.0.0.0"),
			netip.Prefix{},
			false,
		},
		{
			mustAddr("0.0.0.0"),
			mustAddr("0.0.0.0"),
			mustPrefix("0.0.0.0/32"),
			true,
		},
		{
			mustAddr("::"),
			mustAddr("::"),
			mustPrefix("::/128"),
			true,
		},
		{
			mustAddr("0.0.0.0"),
			mustAddr("0.0.0.5"),
			netip.Prefix{},
			false,
		},
		{
			mustAddr("::"),
			mustAddr("::5"),
			netip.Prefix{},
			false,
		},
		{
			mustAddr("0.0.0.0"),
			mustAddr("0.0.0.3"),
			mustPrefix("0.0.0.0/30"),
			true,
		},
		{
			mustAddr("0.0.0.0"),
			mustAddr("255.255.255.255"),
			mustPrefix("0.0.0.0/0"),
			true,
		},
		{
			mustAddr("10.0.0.0"),
			mustAddr("10.255.255.255"),
			mustPrefix("10.0.0.0/8"),
			true,
		},
		{
			mustAddr("172.16.0.0"),
			mustAddr("172.31.255.255"),
			mustPrefix("172.16.0.0/12"),
			true,
		},
		{
			mustAddr("::ffff:0.0.0.0"),
			mustAddr("::ffff:255.255.255.255"),
			mustPrefix("::ffff:0.0.0.0/96"),
			true,
		},
		{
			mustAddr("::"),
			mustAddr("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
			mustPrefix("::/0"),
			true,
		},
		{
			mustAddr("fe80::"),
			mustAddr("febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
			mustPrefix("fe80::/10"),
			true,
		},
	}

	for _, tt := range tests {
		p, ok := extnetip.Prefix(tt.ip1, tt.ip2)
		if ok != tt.ok {
			t.Fatalf("Prefix(%s, %s), got ok: %v, expected: %v", tt.ip1, tt.ip2, ok, tt.ok)
		}
		if p != tt.p {
			t.Fatalf("Prefix(%s, %s), got prefix: %s, expected: %s", tt.ip1, tt.ip2, p, tt.p)
		}
	}
}

func TestPrefixes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		first netip.Addr
		last  netip.Addr
		want  []netip.Prefix
	}{
		{netip.Addr{}, netip.Addr{}, nil},                      // invalid addrs
		{mustAddr("0.0.0.1"), mustAddr("0.0.0.0"), nil},        // wrong order
		{mustAddr("0.0.0.1"), mustAddr("::1"), nil},            // wrong versions
		{mustAddr("0.0.0.1"), mustAddr("::ffff:1.2.3.4"), nil}, // wrong versions

		{mustAddr("0.0.0.0"), mustAddr("255.255.255.255"), pfxSlice("0.0.0.0/0")},
		{mustAddr("::"), mustAddr("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), pfxSlice("::/0")},
		{mustAddr("::ffff:0.0.0.0"), mustAddr("::ffff:255.255.255.255"), pfxSlice("::ffff:0.0.0.0/96")},

		{mustAddr("10.0.0.0"), mustAddr("10.255.255.255"), pfxSlice("10.0.0.0/8")},
		{mustAddr("10.0.0.0"), mustAddr("10.127.255.255"), pfxSlice("10.0.0.0/9")},
		{mustAddr("0.0.0.4"), mustAddr("0.0.0.11"), pfxSlice("0.0.0.4/30", "0.0.0.8/30")},
		{mustAddr("10.0.0.0"), mustAddr("11.10.255.255"), pfxSlice("10.0.0.0/8", "11.0.0.0/13", "11.8.0.0/15", "11.10.0.0/16")},
		{mustAddr("fe80::"), mustAddr("fe80::8"), pfxSlice("fe80::/125", "fe80::8/128")},

		{mustAddr("0.0.0.1"), mustAddr("255.255.255.254"), pfxSlice(
			"0.0.0.1/32",
			"0.0.0.2/31",
			"0.0.0.4/30",
			"0.0.0.8/29",
			"0.0.0.16/28",
			"0.0.0.32/27",
			"0.0.0.64/26",
			"0.0.0.128/25",
			"0.0.1.0/24",
			"0.0.2.0/23",
			"0.0.4.0/22",
			"0.0.8.0/21",
			"0.0.16.0/20",
			"0.0.32.0/19",
			"0.0.64.0/18",
			"0.0.128.0/17",
			"0.1.0.0/16",
			"0.2.0.0/15",
			"0.4.0.0/14",
			"0.8.0.0/13",
			"0.16.0.0/12",
			"0.32.0.0/11",
			"0.64.0.0/10",
			"0.128.0.0/9",
			"1.0.0.0/8",
			"2.0.0.0/7",
			"4.0.0.0/6",
			"8.0.0.0/5",
			"16.0.0.0/4",
			"32.0.0.0/3",
			"64.0.0.0/2",
			"128.0.0.0/2",
			"192.0.0.0/3",
			"224.0.0.0/4",
			"240.0.0.0/5",
			"248.0.0.0/6",
			"252.0.0.0/7",
			"254.0.0.0/8",
			"255.0.0.0/9",
			"255.128.0.0/10",
			"255.192.0.0/11",
			"255.224.0.0/12",
			"255.240.0.0/13",
			"255.248.0.0/14",
			"255.252.0.0/15",
			"255.254.0.0/16",
			"255.255.0.0/17",
			"255.255.128.0/18",
			"255.255.192.0/19",
			"255.255.224.0/20",
			"255.255.240.0/21",
			"255.255.248.0/22",
			"255.255.252.0/23",
			"255.255.254.0/24",
			"255.255.255.0/25",
			"255.255.255.128/26",
			"255.255.255.192/27",
			"255.255.255.224/28",
			"255.255.255.240/29",
			"255.255.255.248/30",
			"255.255.255.252/31",
			"255.255.255.254/32",
		)},

		{mustAddr("::1"), mustAddr("ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe"), pfxSlice(
			"::1/128",
			"::2/127",
			"::4/126",
			"::8/125",
			"::10/124",
			"::20/123",
			"::40/122",
			"::80/121",
			"::100/120",
			"::200/119",
			"::400/118",
			"::800/117",
			"::1000/116",
			"::2000/115",
			"::4000/114",
			"::8000/113",
			"::1:0/112",
			"::2:0/111",
			"::4:0/110",
			"::8:0/109",
			"::10:0/108",
			"::20:0/107",
			"::40:0/106",
			"::80:0/105",
			"::100:0/104",
			"::200:0/103",
			"::400:0/102",
			"::800:0/101",
			"::1000:0/100",
			"::2000:0/99",
			"::4000:0/98",
			"::8000:0/97",
			"::1:0:0/96",
			"::2:0:0/95",
			"::4:0:0/94",
			"::8:0:0/93",
			"::10:0:0/92",
			"::20:0:0/91",
			"::40:0:0/90",
			"::80:0:0/89",
			"::100:0:0/88",
			"::200:0:0/87",
			"::400:0:0/86",
			"::800:0:0/85",
			"::1000:0:0/84",
			"::2000:0:0/83",
			"::4000:0:0/82",
			"::8000:0:0/81",
			"::1:0:0:0/80",
			"::2:0:0:0/79",
			"::4:0:0:0/78",
			"::8:0:0:0/77",
			"::10:0:0:0/76",
			"::20:0:0:0/75",
			"::40:0:0:0/74",
			"::80:0:0:0/73",
			"::100:0:0:0/72",
			"::200:0:0:0/71",
			"::400:0:0:0/70",
			"::800:0:0:0/69",
			"::1000:0:0:0/68",
			"::2000:0:0:0/67",
			"::4000:0:0:0/66",
			"::8000:0:0:0/65",
			"0:0:0:1::/64",
			"0:0:0:2::/63",
			"0:0:0:4::/62",
			"0:0:0:8::/61",
			"0:0:0:10::/60",
			"0:0:0:20::/59",
			"0:0:0:40::/58",
			"0:0:0:80::/57",
			"0:0:0:100::/56",
			"0:0:0:200::/55",
			"0:0:0:400::/54",
			"0:0:0:800::/53",
			"0:0:0:1000::/52",
			"0:0:0:2000::/51",
			"0:0:0:4000::/50",
			"0:0:0:8000::/49",
			"0:0:1::/48",
			"0:0:2::/47",
			"0:0:4::/46",
			"0:0:8::/45",
			"0:0:10::/44",
			"0:0:20::/43",
			"0:0:40::/42",
			"0:0:80::/41",
			"0:0:100::/40",
			"0:0:200::/39",
			"0:0:400::/38",
			"0:0:800::/37",
			"0:0:1000::/36",
			"0:0:2000::/35",
			"0:0:4000::/34",
			"0:0:8000::/33",
			"0:1::/32",
			"0:2::/31",
			"0:4::/30",
			"0:8::/29",
			"0:10::/28",
			"0:20::/27",
			"0:40::/26",
			"0:80::/25",
			"0:100::/24",
			"0:200::/23",
			"0:400::/22",
			"0:800::/21",
			"0:1000::/20",
			"0:2000::/19",
			"0:4000::/18",
			"0:8000::/17",
			"1::/16",
			"2::/15",
			"4::/14",
			"8::/13",
			"10::/12",
			"20::/11",
			"40::/10",
			"80::/9",
			"100::/8",
			"200::/7",
			"400::/6",
			"800::/5",
			"1000::/4",
			"2000::/3",
			"4000::/2",
			"8000::/2",
			"c000::/3",
			"e000::/4",
			"f000::/5",
			"f800::/6",
			"fc00::/7",
			"fe00::/8",
			"ff00::/9",
			"ff80::/10",
			"ffc0::/11",
			"ffe0::/12",
			"fff0::/13",
			"fff8::/14",
			"fffc::/15",
			"fffe::/16",
			"ffff::/17",
			"ffff:8000::/18",
			"ffff:c000::/19",
			"ffff:e000::/20",
			"ffff:f000::/21",
			"ffff:f800::/22",
			"ffff:fc00::/23",
			"ffff:fe00::/24",
			"ffff:ff00::/25",
			"ffff:ff80::/26",
			"ffff:ffc0::/27",
			"ffff:ffe0::/28",
			"ffff:fff0::/29",
			"ffff:fff8::/30",
			"ffff:fffc::/31",
			"ffff:fffe::/32",
			"ffff:ffff::/33",
			"ffff:ffff:8000::/34",
			"ffff:ffff:c000::/35",
			"ffff:ffff:e000::/36",
			"ffff:ffff:f000::/37",
			"ffff:ffff:f800::/38",
			"ffff:ffff:fc00::/39",
			"ffff:ffff:fe00::/40",
			"ffff:ffff:ff00::/41",
			"ffff:ffff:ff80::/42",
			"ffff:ffff:ffc0::/43",
			"ffff:ffff:ffe0::/44",
			"ffff:ffff:fff0::/45",
			"ffff:ffff:fff8::/46",
			"ffff:ffff:fffc::/47",
			"ffff:ffff:fffe::/48",
			"ffff:ffff:ffff::/49",
			"ffff:ffff:ffff:8000::/50",
			"ffff:ffff:ffff:c000::/51",
			"ffff:ffff:ffff:e000::/52",
			"ffff:ffff:ffff:f000::/53",
			"ffff:ffff:ffff:f800::/54",
			"ffff:ffff:ffff:fc00::/55",
			"ffff:ffff:ffff:fe00::/56",
			"ffff:ffff:ffff:ff00::/57",
			"ffff:ffff:ffff:ff80::/58",
			"ffff:ffff:ffff:ffc0::/59",
			"ffff:ffff:ffff:ffe0::/60",
			"ffff:ffff:ffff:fff0::/61",
			"ffff:ffff:ffff:fff8::/62",
			"ffff:ffff:ffff:fffc::/63",
			"ffff:ffff:ffff:fffe::/64",
			"ffff:ffff:ffff:ffff::/65",
			"ffff:ffff:ffff:ffff:8000::/66",
			"ffff:ffff:ffff:ffff:c000::/67",
			"ffff:ffff:ffff:ffff:e000::/68",
			"ffff:ffff:ffff:ffff:f000::/69",
			"ffff:ffff:ffff:ffff:f800::/70",
			"ffff:ffff:ffff:ffff:fc00::/71",
			"ffff:ffff:ffff:ffff:fe00::/72",
			"ffff:ffff:ffff:ffff:ff00::/73",
			"ffff:ffff:ffff:ffff:ff80::/74",
			"ffff:ffff:ffff:ffff:ffc0::/75",
			"ffff:ffff:ffff:ffff:ffe0::/76",
			"ffff:ffff:ffff:ffff:fff0::/77",
			"ffff:ffff:ffff:ffff:fff8::/78",
			"ffff:ffff:ffff:ffff:fffc::/79",
			"ffff:ffff:ffff:ffff:fffe::/80",
			"ffff:ffff:ffff:ffff:ffff::/81",
			"ffff:ffff:ffff:ffff:ffff:8000::/82",
			"ffff:ffff:ffff:ffff:ffff:c000::/83",
			"ffff:ffff:ffff:ffff:ffff:e000::/84",
			"ffff:ffff:ffff:ffff:ffff:f000::/85",
			"ffff:ffff:ffff:ffff:ffff:f800::/86",
			"ffff:ffff:ffff:ffff:ffff:fc00::/87",
			"ffff:ffff:ffff:ffff:ffff:fe00::/88",
			"ffff:ffff:ffff:ffff:ffff:ff00::/89",
			"ffff:ffff:ffff:ffff:ffff:ff80::/90",
			"ffff:ffff:ffff:ffff:ffff:ffc0::/91",
			"ffff:ffff:ffff:ffff:ffff:ffe0::/92",
			"ffff:ffff:ffff:ffff:ffff:fff0::/93",
			"ffff:ffff:ffff:ffff:ffff:fff8::/94",
			"ffff:ffff:ffff:ffff:ffff:fffc::/95",
			"ffff:ffff:ffff:ffff:ffff:fffe::/96",
			"ffff:ffff:ffff:ffff:ffff:ffff::/97",
			"ffff:ffff:ffff:ffff:ffff:ffff:8000:0/98",
			"ffff:ffff:ffff:ffff:ffff:ffff:c000:0/99",
			"ffff:ffff:ffff:ffff:ffff:ffff:e000:0/100",
			"ffff:ffff:ffff:ffff:ffff:ffff:f000:0/101",
			"ffff:ffff:ffff:ffff:ffff:ffff:f800:0/102",
			"ffff:ffff:ffff:ffff:ffff:ffff:fc00:0/103",
			"ffff:ffff:ffff:ffff:ffff:ffff:fe00:0/104",
			"ffff:ffff:ffff:ffff:ffff:ffff:ff00:0/105",
			"ffff:ffff:ffff:ffff:ffff:ffff:ff80:0/106",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffc0:0/107",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffe0:0/108",
			"ffff:ffff:ffff:ffff:ffff:ffff:fff0:0/109",
			"ffff:ffff:ffff:ffff:ffff:ffff:fff8:0/110",
			"ffff:ffff:ffff:ffff:ffff:ffff:fffc:0/111",
			"ffff:ffff:ffff:ffff:ffff:ffff:fffe:0/112",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:0/113",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:8000/114",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:c000/115",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:e000/116",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:f000/117",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:f800/118",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fc00/119",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fe00/120",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ff00/121",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ff80/122",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffc0/123",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffe0/124",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fff0/125",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fff8/126",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffc/127",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe/128",
		)},
	}
	for _, tt := range tests {
		got := extnetip.Prefixes(tt.first, tt.last)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("failed %s->%s. got:", tt.first, tt.last)
			for _, v := range got {
				t.Errorf("  %v", v)
			}
			t.Error("want:\n")
			for _, v := range tt.want {
				t.Errorf("  %v", v)
			}
		}
	}
}
