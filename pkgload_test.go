package pkgload

import (
	"errors"
	"fmt"
	"go/build"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestVisitUnits(t *testing.T) {
	tests := []struct {
		path string
		desc string
	}{
		{"./testdata/all_included", "Base+Test+ExternalTest+TestBinary"},
		{"./testdata/base_only", "Base"},
		{"./testdata/base_with_ext_tests", "Base+ExternalTest+TestBinary"},
		{"./testdata/base_with_tests", "Base+Test+TestBinary"},
		{"./testdata/empty", ""},
		{"./testdata/main_only", "Base"},
		{"./testdata/main_with_tests", "Base+Test+TestBinary"},

		{"./testdata/horrors/base-only/blah.test", "Base"},
		{"./testdata/horrors/base-only/blah_test", "Base"},
		{"./testdata/horrors/base-with-tests/blah.test", "Base+Test+TestBinary"},
		{"./testdata/horrors/base-with-tests/blah_test", "Base+Test+TestBinary"},
		{"./testdata/horrors/test_data", "ExternalTest"},
	}

	checkFields := func(desc string, u *Unit) error {
		for _, key := range strings.Split(desc, "+") {
			switch key {
			case "Base":
				if u.Base == nil {
					return errors.New("Base is missing")
				}
			case "Test":
				if u.Test == nil {
					return errors.New("Test is missing")
				}
			case "ExternalTest":
				if u.ExternalTest == nil {
					return errors.New("ExternalTest is missing")
				}
			case "TestBinary":
				if u.TestBinary == nil {
					return errors.New("TestBinary is missing")
				}
			default:
				panic(fmt.Sprintf("unexpected key: %v", key))
			}
		}
		return nil
	}

	type testPackage struct {
		filePath string
		pkgPath  string
		desc     string
	}

	paths := make([]string, len(tests))
	testsMap := make(map[string]testPackage)
	for i, test := range tests {
		pkgPath := "github.com/go-toolsmith/pkgload/" + strings.TrimPrefix(test.path, "./")
		paths[i] = tests[i].path
		absPath, err := filepath.Abs(tests[i].path)
		if err != nil {
			t.Fatalf("get abs path: %v", err)
		}
		testsMap[pkgPath] = testPackage{
			filePath: absPath,
			pkgPath:  pkgPath,
			desc:     test.desc,
		}
	}

	runWithMode := func(name string, mode packages.LoadMode, fn func(*packages.Config, *testing.T)) {
		t.Run(name, func(t *testing.T) {
			cfg := packages.Config{Mode: mode, Tests: true, Fset: token.NewFileSet()}
			fn(&cfg, t)
		})
	}
	runWithAllModes := func(name string, fn func(*packages.Config, *testing.T)) {
		runWithMode(name+"/Files", packages.LoadFiles, fn)
		runWithMode(name+"/LoadImports", packages.LoadImports, fn)
		runWithMode(name+"/LoadTypes", packages.LoadTypes, fn)
		runWithMode(name+"/LoadSyntax", packages.LoadSyntax, fn)
	}

	// Check that loading GOROOT packages does not cause
	// VisitUnits to panic.
	runWithAllModes("loadStd", func(cfg *packages.Config, t *testing.T) {
		goroot := build.Default.GOROOT
		wd, err := os.Getwd()
		if err != nil {
			t.Skipf("can't get wd: %v", err)
		}
		defer func(prev string) {
			if err := os.Chdir(prev); err != nil {
				panic(fmt.Sprintf("can't go back: %v", err))
			}
		}(wd)
		if err := os.Chdir(goroot); err != nil {
			t.Skipf("chdir: %v", err)
		}
		pkgs, err := packages.Load(cfg, "std")
		if err != nil {
			t.Fatalf("load packages: %v", err)
		}
		VisitUnits(pkgs, func(u *Unit) {})
	})

	runWithAllModes("loadAll", func(cfg *packages.Config, t *testing.T) {
		pkgs, err := packages.Load(cfg, paths...)
		if err != nil {
			t.Fatalf("load packages: %v", err)
		}
		remains := len(testsMap) - 1 // Substract the empty unit
		VisitUnits(pkgs, func(u *Unit) {
			p, ok := testsMap[u.NonNil().PkgPath]
			if !ok {
				t.Fatalf("unmatched pkg path %q", u.NonNil().PkgPath)
			}
			remains--
			if err := checkFields(p.desc, u); err != nil {
				t.Errorf("%q: check %q: %v",
					u.NonNil().PkgPath, p.desc, err)
			}
		})
		if remains != 0 {
			t.Errorf("unprocessed units: %d", remains)
		}
	})

	runWithAllModes("loadOneByOne", func(cfg *packages.Config, t *testing.T) {
		for _, path := range paths {
			pkgs, err := packages.Load(cfg, path)
			if err != nil {
				t.Fatalf("load packages: %v", err)
			}
			VisitUnits(pkgs, func(u *Unit) {
				p, ok := testsMap[u.NonNil().PkgPath]
				if !ok {
					t.Fatalf("unmatched pkg path %q", u.NonNil().PkgPath)
				}
				if err := checkFields(p.desc, u); err != nil {
					t.Errorf("%q: check %q: %v",
						u.NonNil().PkgPath, p.desc, err)
				}
			})
		}
	})
}
