# package extnetip
[![Go Reference](https://pkg.go.dev/badge/github.com/gaissmai/extnetip.svg)](https://pkg.go.dev/github.com/gaissmai/extnetip#section-documentation)
[![CI](https://github.com/gaissmai/extnetip/actions/workflows/go.yml/badge.svg)](https://github.com/gaissmai/extnetip/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/gaissmai/extnetip/badge.svg?branch=master)](https://coveralls.io/github/gaissmai/extnetip?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gaissmai/extnetip)](https://goreportcard.com/report/github.com/gaissmai/extnetip)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`package extnetip` is an extension package to the stdlib `net/netip`.

Some missing math functions are added to the closed private internals of `netip.Addr` using unsafe.

No further types are defined, only helper functions on the existing types `netip.Addr` and `netip.Prefix`.

Based on the stdlib `net/netip` and this `extnetip` extension package, third party IP-Range libraries
are possible without further low-level IP maths.

## API

```go
func Range(p netip.Prefix) (first, last netip.Addr)

func Prefix(first, last netip.Addr) (prefix netip.Prefix, ok bool)

func Prefixes(first, last netip.Addr) []netip.Prefix

func AppendPrefixes(dst []netip.Prefix, first, last netip.Addr) []netip.Prefix
```

## Future

Hopefully some day these tiny helper functions are added to `netip` by the stdlib maintainers, currently (2022)
they refused [the proposal](https://github.com/golang/go/issues/53236).

## Importpath

`import "github.com/gaissmai/extnetip"`
