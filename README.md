[![Go Report Card](https://goreportcard.com/badge/github.com/go-toolsmith/pkgload)](https://goreportcard.com/report/github.com/go-toolsmith/pkgload)
[![GoDoc](https://godoc.org/github.com/go-toolsmith/pkgload?status.svg)](https://godoc.org/github.com/go-toolsmith/pkgload)


# pkgload

Package pkgload is a set of utilities for `go/packages` load-related operations.

## Installation:

```bash
go get -v github.com/go-toolsmith/pkgload
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
