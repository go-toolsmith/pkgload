package pkgload

import (
	"errors"
	"fmt"
	"go/token"
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

	paths := make([]string, len(tests))
	testsMap := make(map[string]string)
	for i := range tests {
		paths[i] = tests[i].path
		absPath, err := filepath.Abs(tests[i].path)
		if err != nil {
			t.Fatalf("get abs path: %v", err)
		}
		testsMap["_"+absPath] = tests[i].desc
	}

	t.Run("loadAll", func(t *testing.T) {
		cfg := packages.Config{Mode: packages.LoadSyntax, Tests: true, Fset: token.NewFileSet()}
		pkgs, err := packages.Load(&cfg, paths...)
		if err != nil {
			t.Fatalf("load packages: %v", err)
		}
		remains := len(testsMap) - 1 // Substract the empty unit
		VisitUnits(pkgs, func(u *Unit) {
			desc, ok := testsMap[u.Base.PkgPath]
			if !ok {
				t.Fatalf("unmatched pkg path %q", u.Base.PkgPath)
			}
			remains--
			if err := checkFields(desc, u); err != nil {
				t.Errorf("%q: check %q: %v",
					u.Base.PkgPath, desc, err)
			}
		})
		if remains != 0 {
			t.Errorf("unprocessed units: %d", remains)
		}
	})

	t.Run("loadOneByOne", func(t *testing.T) {
		cfg := packages.Config{Mode: packages.LoadSyntax, Tests: true, Fset: token.NewFileSet()}
		for _, path := range paths {
			pkgs, err := packages.Load(&cfg, path)
			if err != nil {
				t.Fatalf("load packages: %v", err)
			}
			VisitUnits(pkgs, func(u *Unit) {
				desc, ok := testsMap[u.Base.PkgPath]
				if !ok {
					t.Fatalf("unmatched pkg path %q", u.Base.PkgPath)
				}
				if err := checkFields(desc, u); err != nil {
					t.Errorf("%q: check %q: %v",
						u.Base.PkgPath, desc, err)
				}
			})
		}
	})
}
