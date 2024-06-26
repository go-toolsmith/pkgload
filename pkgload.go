// Package pkgload is a set of utilities for `go/packages` load-related operations.
package pkgload

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Unit is a set of packages that form a logical group.
// It is guaranteed that at least 1 field of this object is non-nil.
type Unit struct {
	// Base is a standard (normal) package.
	//
	// Note: it can be nil for packages that only have external
	// tests, for example.
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

// NonNil returns first non-nil field (package) of the unit.
//
//  1. If Base is not nil, return Base.
//  2. If Test is not nil, return Test.
//  3. If ExternalTest is not nil, return ExternalTest.
//  4. Otherwise return TestBinary.
//
// If all unit fields are nil, method panics.
// This should never happen for properly-loaded units.
func (u *Unit) NonNil() *packages.Package {
	switch {
	case u.Base != nil:
		return u.Base
	case u.Test != nil:
		return u.Test
	case u.ExternalTest != nil:
		return u.ExternalTest
	case u.TestBinary != nil:
		return u.TestBinary
	default:
		panic("all Unit fields are nil")
	}
}

// LoadPackages with a given config and patterns.
func LoadPackages(cfg *packages.Config, patterns []string) ([]*packages.Package, error) {
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}

	result := pkgs[:0]
	VisitUnits(pkgs, func(u *Unit) {
		if u.ExternalTest != nil {
			result = append(result, u.ExternalTest)
		}

		switch {
		// Prefer tests to the base package, if present.
		case u.Test != nil:
			result = append(result, u.Test)
		case u.Base != nil:
			result = append(result, u.Base)
		}
	})

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].PkgPath < result[j].PkgPath
	})
	return result, nil
}

// Deduplicate returns a copy of pkgs slice where all duplicated
// package entries are removed.
//
// Packages are considered equal if all conditions below are satisfied:
//   - Same ID
//   - Same Name
//   - Same PkgPath
//   - Equal GoFiles
func Deduplicate(pkgs []*packages.Package) []*packages.Package {
	type pkgKey struct {
		id    string
		name  string
		path  string
		files string
	}

	pkgSet := make(map[pkgKey]*packages.Package)
	for _, pkg := range pkgs {
		sort.Strings(pkg.GoFiles)
		key := pkgKey{
			id:    pkg.ID,
			name:  pkg.Name,
			path:  pkg.PkgPath,
			files: strings.Join(pkg.GoFiles, ";"),
		}
		pkgSet[key] = pkg
	}

	list := make([]*packages.Package, 0, len(pkgSet))
	for _, pkg := range pkgSet {
		list = append(list, pkg)
	}
	return list
}

// VisitUnits traverses potentially unsorted pkgs list as a set of units.
// All related packages from the slice are passed into visit func as a single unit.
// Units are visited in a sorted order (import path).
//
// All packages in a slice must be non-nil.
func VisitUnits(pkgs []*packages.Package, visit func(*Unit)) {
	pkgs = Deduplicate(pkgs)
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
			panic(fmt.Sprintf("nil assertion failed for ID=%q Path=%q",
				pkg.ID, pkg.PkgPath))
		}
	}

	withoutSuffix := func(s, suffix string) string {
		return s[:len(s)-len(suffix)]
	}

	for _, pkg := range pkgs {
		switch {
		case strings.HasSuffix(pkg.Name, "_test"):
			key := withoutSuffix(pkg.PkgPath, "_test")
			u := internUnit(key)
			mustBeNil(u.ExternalTest)
			u.ExternalTest = pkg
		case strings.Contains(pkg.ID, ".test]"):
			u := internUnit(pkg.PkgPath)
			mustBeNil(u.Test)
			u.Test = pkg
		case pkg.Name == "main" && strings.HasSuffix(pkg.ID, ".test"):
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
	sort.SliceStable(unitList, func(i, j int) bool {
		return unitList[i].NonNil().PkgPath < unitList[j].NonNil().PkgPath
	})
	for _, u := range unitList {
		visit(u)
	}
}
