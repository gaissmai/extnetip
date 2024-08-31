# package extnetip
[![Go Reference](https://pkg.go.dev/badge/github.com/gaissmai/extnetip.svg)](https://pkg.go.dev/github.com/gaissmai/extnetip#section-documentation)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/gaissmai/extnetip)
[![CI](https://github.com/gaissmai/extnetip/actions/workflows/go.yml/badge.svg)](https://github.com/gaissmai/extnetip/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gaissmai/extnetip)](https://goreportcard.com/report/github.com/gaissmai/extnetip)
[![Coverage Status](https://coveralls.io/repos/github/gaissmai/extnetip/badge.svg?branch=master)](https://coveralls.io/github/gaissmai/extnetip?branch=master)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Stand With Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/badges/StandWithUkraine.svg)](https://stand-with-ukraine.pp.ua)


Package `extnetip` is an extension to `net/netip` with
a few missing but important auxiliary functions for
converting IP-prefixes to IP-ranges and vice versa.

With these extensions to `net/netip`, third-party IP-range
libraries become easily possible.

## API

```go
import "github.com/gaissmai/extnetip"

func Range(p netip.Prefix) (first, last netip.Addr)
func Prefix(first, last netip.Addr) (prefix netip.Prefix, ok bool)
func All(first, last netip.Addr) iter.Seq[netip.Prefix]

// Deprecated: func Prefixes(first, last netip.Addr) []netip.Prefix
// Deprecated: func PrefixesAppend(dst []netip.Prefix, first, last netip.Addr) []netip.Prefix
```

## Future

Hopefully some day these needed helper functions are added to `netip` by the stdlib maintainers.
