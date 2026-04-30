package main

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "panicfatal",
	Doc:  "проверяет, что нет вызовов panic, нет вызовов log.Fatal/os.Exit вне функции main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	panicCall := func(x *ast.ExprStmt, funcDecl *ast.FuncDecl) {
		if call, ok := x.X.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
				pass.Reportf(ident.NamePos, "call panic")
				return
			}

			if pass.Pkg.Name() == "main" && funcDecl != nil && funcDecl.Name.Name == "main" {
				return
			}

			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				selX, okX := sel.X.(*ast.Ident)
				if !okX {
					return
				}
				if selX.Name == "log" && sel.Sel.Name == "Fatal" {
					pass.Reportf(selX.Pos(), "call log.Fatal")
					return
				}
				if selX.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(selX.Pos(), "call os.Exit")
				}
			}
		}
	}

	var currentFunc *ast.FuncDecl

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			filename := filepath.Base(pass.Fset.Position(file.Pos()).Filename)
			if strings.HasPrefix(filename, "mock_") {
				return true
			}
			switch x := node.(type) {
			case *ast.FuncDecl:
				currentFunc = x
			case *ast.ExprStmt:
				panicCall(x, currentFunc)
			}
			return true
		})
	}

	return nil, nil
}
