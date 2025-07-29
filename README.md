# package extnetip
[![Go Reference](https://pkg.go.dev/badge/github.com/gaissmai/extnetip.svg)](https://pkg.go.dev/github.com/gaissmai/extnetip#section-documentation)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/gaissmai/extnetip)
[![CI](https://github.com/gaissmai/extnetip/actions/workflows/go.yml/badge.svg)](https://github.com/gaissmai/extnetip/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gaissmai/extnetip)](https://goreportcard.com/report/github.com/gaissmai/extnetip)
[![Coverage Status](https://coveralls.io/repos/github/gaissmai/extnetip/badge.svg?branch=master)](https://coveralls.io/github/gaissmai/extnetip?branch=master)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Stand With Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/badges/StandWithUkraine.svg)](https://stand-with-ukraine.pp.ua)


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
func All(first, last netip.Addr) iter.Seq[netip.Prefix]

// Deprecated: func Prefixes(first, last netip.Addr) []netip.Prefix
// Deprecated: func PrefixesAppend(dst []netip.Prefix, first, last netip.Addr) []netip.Prefix
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
cpu: Intel(R) Core(TM) i5-7500T CPU @ 2.70GHz
                   │   safe.bm    │             unsafe.bm              │
                   │    sec/op    │   sec/op     vs base               │
Conversion/peek_v4    3.381n ± 1%   2.461n ± 0%  -27.21% (p=0.002 n=6)
Conversion/peek_v6    4.303n ± 1%   2.463n ± 1%  -42.77% (p=0.002 n=6)
Conversion/back_v4    3.177n ± 0%   2.460n ± 1%  -22.60% (p=0.002 n=6)
Conversion/back_v6    9.557n ± 1%   2.463n ± 1%  -74.23% (p=0.002 n=6)
Range/v4             13.270n ± 0%   5.852n ± 1%  -55.90% (p=0.002 n=6)
Range/v6             27.075n ± 1%   5.987n ± 0%  -77.89% (p=0.002 n=6)
Prefix/v4             17.89n ± 0%   12.97n ± 0%  -27.50% (p=0.002 n=6)
Prefix/v6             19.05n ± 1%   12.11n ± 0%  -36.46% (p=0.002 n=6)
geomean               9.261n        4.604n       -50.28%
```

## Future Work

It is hoped that these frequently needed helper functions will be added to the Go standard
library's `netip` package at some point in the future by the maintainers.

Until then, `extnetip` provides a robust and efficient alternative.
