# package extnetip


`extnetip` is an extension package to the stdlib `net/netip`.

Some missing math functions are added to the closed private internals of `netip.Addr` using unsafe.

No further types are defined, only helper functions on the existing types `netip.Addr` and `netip.Prefix`.

With these tiny extensions, third party IP-Range libraries, based on the
stdlib net/netip, are now possible without further bytes/bits fumbling.


```go
func Range(p netip.Prefix) (first, last netip.Addr)

func Prefix(first, last netip.Addr) (prefix netip.Prefix, ok bool)

func Prefixes(first, last netip.Addr) []netip.Prefix

func AppendPrefixes(dst []netip.Prefix, first, last netip.Addr) []netip.Prefix
```

Hopefully some day in the future this is no longer needed when the go
authors add these tiny missing helper functions to netip, currently (2022)
they refused [the proposal](https://github.com/golang/go/issues/53236).

## Importpath

`import "github.com/gaissmai/extnetip"`
