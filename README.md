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
cpu: Intel(R) Core(TM) i5-8250U CPU @ 1.60GHz
          │   safe.bm    │             unsafe.bm              │
          │    sec/op    │   sec/op     vs base               │
Range/v4    13.440n ± 3%   5.147n ± 3%  -61.71% (p=0.002 n=6)
Range/v6    26.550n ± 0%   5.377n ± 3%  -79.75% (p=0.002 n=6)
Prefix/v4    17.43n ± 2%   12.74n ± 3%  -26.86% (p=0.002 n=6)
Prefix/v6    19.04n ± 4%   11.09n ± 1%  -41.74% (p=0.002 n=6)
All/v4       1.789µ ± 3%   1.509µ ± 4%  -15.65% (p=0.002 n=6)
All/v6       7.851µ ± 1%   5.654µ ± 4%  -27.98% (p=0.002 n=6)
geomean      108.8n        56.74n       -47.87%
```

## Future Work

It is hoped that these frequently needed helper functions will be added to the Go standard
library's `netip` package at some point in the future by the maintainers.

Until then, `extnetip` provides a robust and efficient alternative.
