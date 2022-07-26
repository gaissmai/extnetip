# package extnetip
[![Go Reference](https://pkg.go.dev/badge/github.com/gaissmai/extnetip.svg)](https://pkg.go.dev/github.com/gaissmai/extnetip#section-documentation)
[![CI](https://github.com/gaissmai/extnetip/actions/workflows/go.yml/badge.svg)](https://github.com/gaissmai/extnetip/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/gaissmai/extnetip/badge.svg?branch=master)](https://coveralls.io/github/gaissmai/extnetip?branch=master)
[![Stand With Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/badges/StandWithUkraine.svg)](https://stand-with-ukraine.pp.ua)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`package extnetip` is an extension to `net/netip`.

No additional types are defined, only required auxiliary functions for some
existing `net/netip` types are provided.

With these small extensions, third-party IP range libraries based on stdlib
`net/netip` are now possible without frequent conversion to/from bytes, see also https://github.com/gaissmai/iprange


## API

```go
import "github.com/gaissmai/extnetip"

func Range(p netip.Prefix) (first, last netip.Addr)

func Prefix(first, last netip.Addr) (prefix netip.Prefix, ok bool)

func Prefixes(first, last netip.Addr) []netip.Prefix

func AppendPrefixes(dst []netip.Prefix, first, last netip.Addr) []netip.Prefix
```

## Future

Hopefully some day these tiny helper functions are added to `netip` by the stdlib maintainers, currently (2022)
they refused [the proposal](https://github.com/golang/go/issues/53236).
