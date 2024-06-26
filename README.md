# pkgload

[![build-img]][build-url]
[![pkg-img]][pkg-url]
[![reportcard-img]][reportcard-url]
[![version-img]][version-url]

Package `pkgload` is a set of utilities for `go/packages` load-related operations.

## Installation:

Go version 1.17+

```bash
go get github.com/go-toolsmith/pkgload
```

## Example

```go
package main

import (
	"fmt"
	"go/token"

	"github.com/go-toolsmith/pkgload"
	"golang.org/x/tools/go/packages"
)

func main() {
	fset := token.NewFileSet()
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: true,
		Fset:  fset,
	}

	patterns := []string{"mypackage"}
	pkgs, err := pkgload.LoadPackages(cfg, patterns)
	if err != nil {
		panic(err)
	}

	pkgs = pkgload.Deduplicate(pkgs)

	pkgload.VisitUnits(pkgs, func(u *pkgload.Unit) {
		pkgPath := u.NonNil().PkgPath
		println(pkgPath)
	})
}
```

## License

[MIT License](LICENSE).

[build-img]: https://github.com/go-toolsmith/pkgload/workflows/build/badge.svg
[build-url]: https://github.com/go-toolsmith/pkgload/actions
[pkg-img]: https://pkg.go.dev/badge/go-toolsmith/pkgload
[pkg-url]: https://pkg.go.dev/github.com/go-toolsmith/pkgload
[reportcard-img]: https://goreportcard.com/badge/go-toolsmith/pkgload
[reportcard-url]: https://goreportcard.com/report/go-toolsmith/pkgload
[version-img]: https://img.shields.io/github/v/release/go-toolsmith/pkgload
[version-url]: https://github.com/go-toolsmith/pkgload/releases
