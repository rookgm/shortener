// Package main includes the standard static and staticcheck analyzers.
//
// Standard static analyzers: printf, shadow, structtag.
//
// The SA category of checks, codenamed staticcheck,
// that are concerned with the correctness of code.
//
// The S category of checks, codenamed simple,
// contains S1000, S1001, S1002, S1008, S1011 checks that are
// concerned with simplifying code.
//
// S1000 - use plain channel send or receive instead of single-case select;
//
// S1001 - replace for loop with call to copy;
//
// S1002 - omit comparison with boolean constant;
//
// S1008 - simplify returning boolean expression;
//
// S1011 - use a single append to concatenate two slices.
//
// The ST category of checks, codenamed stylecheck,
// contains ST1001, ST1006, ST1012 checks that are
// concerned with stylistic issues.
//
// ST1001 - dot imports are discouraged;
//
// ST1006 - poorly chosen receiver name;
//
// ST1012 - poorly chosen name for error variable.
//
// The QF category of checks, codenamed quickfix,
// contains QF1002, QF1005, QF1007 checks that are
// used as part of gopls for automatic refactorings.
//
// QF1002 - convert untagged switch to tagged switch;
//
// QF1005 - expand call to math.Pow;
//
// QF1007 - merge conditional assignment into variable declaration;
//
// Analyzer osexit that checks for call os.Exit in
// main function of package main.go
//
// # Usage
//
// go run cmd/staticlint/main.go ./...
package main
