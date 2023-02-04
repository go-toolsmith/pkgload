# pkgload

[![build-img]][build-url]
[![pkg-img]][pkg-url]
[![reportcard-img]][reportcard-url]
[![coverage-img]][coverage-url]
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
	cfg := packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: true,
		Fset:  fset,
	}

	patterns := []string{"mypackage"}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}

	result := pkgs[:0]
	pkgload.VisitUnits(pkgs, func(u *pkgload.Unit) {
		if u.ExternalTest != nil {
			result = append(result, u.ExternalTest)
		}
		result = append(result, u.Base)
	})
}
```

## License

[MIT License](LICENSE).

[build-img]: https://github.com/go-toolsmith/astp/workflows/build/badge.svg
[build-url]: https://github.com/go-toolsmith/astp/actions
[pkg-img]: https://pkg.go.dev/badge/go-toolsmith/astp
[pkg-url]: https://pkg.go.dev/github.com/go-toolsmith/astp
[reportcard-img]: https://goreportcard.com/badge/go-toolsmith/astp
[reportcard-url]: https://goreportcard.com/report/go-toolsmith/astp
[coverage-img]: https://codecov.io/gh/go-toolsmith/astp/branch/main/graph/badge.svg
[coverage-url]: https://codecov.io/gh/go-toolsmith/astp
[version-img]: https://img.shields.io/github/v/release/go-toolsmith/astp
[version-url]: https://github.com/go-toolsmith/astp/releases
