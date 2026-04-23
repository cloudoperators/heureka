// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

// This is a helper command intended for use within Heureka.
// It verifies that all calls to the cache.CallCached function are made correctly.
// Specifically, it ensures that the function name string passed to CallCached
// matches the actual function being cached and that the arguments are appropriate.
//
// This helps prevent rare, hard-to-detect bugs where an incorrect function name
// could lead to cache collisions or data leakage, posing a potential security risk.
//
// This tool is designed to be invoked via the `go generate` command.

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type result struct {
	pos       token.Position
	want, got string
}

// extractFuncName unwraps expressions like wrappers and returns the underlying function name.
func extractFuncName(expr ast.Expr) (string, bool) {
	switch fn := expr.(type) {
	case *ast.Ident:
		return fn.Name, true

	case *ast.SelectorExpr:
		return fn.Sel.Name, true

	case *ast.CallExpr:
		// Assume the actual function is the last argument
		if len(fn.Args) == 0 {
			return "", false
		}

		return extractFuncName(fn.Args[len(fn.Args)-1])
	}

	return "", false
}

func main() {
	dir := flag.String("dir", ".", "directory to scan (walked recursively)")

	flag.Parse()

	var mismatches []result

	fset := token.NewFileSet()

	err := filepath.WalkDir(*dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		ast.Inspect(node, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Unwrap generics if present
			funExpr := call.Fun
			if idx, ok := funExpr.(*ast.IndexExpr); ok {
				funExpr = idx.X
			}

			var isCallCached bool

			switch fn := funExpr.(type) {
			case *ast.Ident:
				isCallCached = fn.Name == "CallCached"
			case *ast.SelectorExpr:
				isCallCached = fn.Sel.Name == "CallCached"
			}

			if !isCallCached {
				return true
			}

			if len(call.Args) < 4 {
				mismatches = append(mismatches, result{
					pos:  fset.Position(call.Lparen),
					want: "at least 4 arguments",
					got:  fmt.Sprintf("%d arguments", len(call.Args)),
				})

				return true
			}

			// 3rd argument: function name string
			cacheFuncNameArg := call.Args[2]

			strLit, ok := cacheFuncNameArg.(*ast.BasicLit)
			if !ok || strLit.Kind != token.STRING {
				mismatches = append(mismatches, result{
					pos:  fset.Position(cacheFuncNameArg.Pos()),
					want: "string literal",
					got:  fmt.Sprintf("%T", cacheFuncNameArg),
				})

				return true
			}

			funcNameStr := strings.Trim(strLit.Value, `"`)

			cacheFunc := call.Args[3]

			funcName, ok := extractFuncName(cacheFunc)
			if !ok {
				mismatches = append(mismatches, result{
					pos:  fset.Position(cacheFunc.Pos()),
					want: "callable function",
					got:  fmt.Sprintf("%T", cacheFunc),
				})

				return true
			}

			if funcNameStr != funcName {
				mismatches = append(mismatches, result{
					pos:  fset.Position(strLit.Pos()),
					want: funcName,
					got:  funcNameStr,
				})
			}

			return true
		})

		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan error:", err)
		os.Exit(1)
	}

	for _, m := range mismatches {
		fmt.Fprintf(os.Stderr, "%s: mismatch - want %q, got %q\n",
			m.pos, m.want, m.got)
	}

	if len(mismatches) > 0 {
		os.Exit(1)
	}
}
