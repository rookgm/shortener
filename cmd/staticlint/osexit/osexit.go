// Package osexit defines an Analyzer that checks direct
// call os.Exit function in main function of package main
package osexit

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"strings"
)

// Analyzer is osexit analyzer
var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check for call os.Exit in main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// find only package "main"
		if strings.Compare(file.Name.Name, "main") != 0 {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if funcDecl, ok := node.(*ast.FuncDecl); ok {
				// find only function "main"
				if strings.Compare(funcDecl.Name.Name, "main") != 0 {
					return true
				}
			}
			if callExpr, ok := node.(*ast.CallExpr); ok {
				switch fun := callExpr.Fun.(type) {
				case *ast.SelectorExpr:
					if selIdent, ok := fun.X.(*ast.Ident); ok {
						if strings.Compare(selIdent.Name, "os") == 0 && strings.Compare(fun.Sel.Name, "Exit") == 0 {
							pass.Reportf(selIdent.NamePos, "call function os.Exit")
						}
					}
				}
			}
			return true
		})
	}

	return nil, nil
}
