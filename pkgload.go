// Package pkgload is a set of utilities for `go/packages` load-related operations.
package pkgload

import (
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Unit is a set of packages that form a logical group.
type Unit struct {
	// Base is a standard (normal) package.
	Base *packages.Package

	// Test is a package compiled for test.
	// Can be nil.
	Test *packages.Package

	// ExternalTest is a "_test" compiled package.
	// Can be nil.
	ExternalTest *packages.Package

	// TestBinary is a test binary.
	// Non-nil if Test or ExternalTest are present.
	TestBinary *packages.Package
}

// VisitUnits traverses potentially unsorted pkgs list as a set of units.
// All related packages from the slice are passed into visit func as a single unit.
// Units are visited in a sorted order (import path).
//
// All packages in a slice must be non-nil.
func VisitUnits(pkgs []*packages.Package, visit func(*Unit)) {
	units := make(map[string]*Unit)

	internUnit := func(key string) *Unit {
		u, ok := units[key]
		if !ok {
			u = &Unit{}
			units[key] = u
		}
		return u
	}

	// Sanity check.
	// Panic should never trigger if this library is correct.
	mustBeNil := func(pkg *packages.Package) {
		if pkg != nil {
			panic("nil assertion failed")
		}
	}

	withoutSuffix := func(s, suffix string) string {
		return s[:len(s)-len(suffix)]
	}

	for _, pkg := range pkgs {
		switch {
		case strings.HasSuffix(pkg.PkgPath, "_test"):
			key := withoutSuffix(pkg.PkgPath, "_test")
			u := internUnit(key)
			mustBeNil(u.ExternalTest)
			u.ExternalTest = pkg
		case strings.Contains(pkg.ID, ".test]"):
			u := internUnit(pkg.PkgPath)
			mustBeNil(u.Test)
			u.Test = pkg
		case pkg.Name == "main" && strings.HasSuffix(pkg.PkgPath, ".test"):
			key := withoutSuffix(pkg.PkgPath, ".text")
			u := internUnit(key)
			mustBeNil(u.TestBinary)
			u.TestBinary = pkg
		case pkg.Name == "":
			// Empty package. Skip.
		default:
			u := internUnit(pkg.PkgPath)
			mustBeNil(u.Base)
			u.Base = pkg
		}
	}

	unitList := make([]*Unit, 0, len(units))
	for _, u := range units {
		unitList = append(unitList, u)
	}
	sort.Slice(unitList, func(i, j int) bool {
		return unitList[i].Base.PkgPath < unitList[j].Base.PkgPath
	})
	for _, u := range unitList {
		visit(u)
	}
}
