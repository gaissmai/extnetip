# package extnetip
[![Go Reference](https://pkg.go.dev/badge/github.com/gaissmai/extnetip.svg)](https://pkg.go.dev/github.com/gaissmai/extnetip#section-documentation)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`extnetip` is an extension package to the stdlib `net/netip`.

Some missing math functions are added to the closed private internals of `netip.Addr` using unsafe.

No further types are defined, only helper functions on the existing types `netip.Addr` and `netip.Prefix`.

With these tiny extensions, third party IP-Range libraries, based on the
stdlib net/netip, are now possible without further bytes/bits fumbling.

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
