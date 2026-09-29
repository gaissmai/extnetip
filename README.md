[![Go Reference](https://pkg.go.dev/badge/github.com/gaissmai/extnetip.svg)](https://pkg.go.dev/github.com/gaissmai/extnetip#section-documentation)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/gaissmai/extnetip)
[![CI](https://github.com/gaissmai/extnetip/actions/workflows/go.yml/badge.svg)](https://github.com/gaissmai/extnetip/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/gaissmai/extnetip/badge.svg?branch=master)](https://coveralls.io/github/gaissmai/extnetip?branch=master)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Stand With Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/badges/StandWithUkraine.svg)](https://stand-with-ukraine.pp.ua)

# package extnetip

Package `extnetip` is an extension to the Go standard library package `net/netip`, providing
a set of important auxiliary functions **currently missing** from `netip` for
converting IP prefixes to IP ranges and vice versa.

With these extensions, it becomes straightforward to build third-party IP-range
libraries based on `net/netip`.

## API

```go
import "github.com/gaissmai/extnetip"

func Range(p netip.Prefix) (first, last netip.Addr)
func Prefix(first, last netip.Addr) (prefix netip.Prefix, ok bool)
func CommonPrefix(pfx1, pfx2 netip.Prefix) (pfx netip.Prefix)
func All(first, last netip.Addr) iter.Seq[netip.Prefix]
```

## Unsafe Mode

This package supports two modes of operation for converting between `netip.Addr` and
a custom internal `uint128` representation:

- When built with the `unsafe` build tag (e.g., `go build -tags=unsafe`), conversions use
  `unsafe.Pointer` to perform zero-copy, direct memory reinterpretation. This method
  is **significantly faster**.

- Without the `unsafe` build tag, conversions are performed safely by using
  `binary.ByteOrder`-based byte slice manipulations, avoiding use of the `unsafe` package.
  This is the default mode and is suitable when importing unsafe modules is prohibited.

### Performance Benchmark

Below is a benchmark comparing the safe (default) and unsafe conversion methods:

```
goos: linux
goarch: amd64
pkg: github.com/gaissmai/extnetip
cpu: AMD Ryzen 7 PRO 4750U with Radeon Graphics
          │   safe.bm    │              unsafe.bm              │
          │    sec/op    │    sec/op     vs base               │
Range/v4    12.410n ± 1%   4.115n ± 32%  -66.84% (p=0.002 n=6)
Range/v6    23.020n ± 1%   4.221n ± 31%  -81.66% (p=0.002 n=6)
Prefix/v4   19.975n ± 1%   9.660n ±  2%  -51.64% (p=0.002 n=6)
Prefix/v6   25.100n ± 1%   9.534n ±  0%  -62.02% (p=0.002 n=6)
geomean      19.45n        6.324n        -67.49%
```

## Future Work

It is hoped that these frequently needed helper functions will be added to the Go standard
library's `netip` package at some point in the future by the maintainers.

Until then, `extnetip` provides a robust and efficient alternative.
