package main

import (
	"encoding/json"
	"github.com/rookgm/shortener/cmd/staticlint/osexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"log"
	"os"
	"strings"
)

// ConfigFileName is config filename contains some checks in the following
// categories: Simple, Stylecheck, Quickfix.
const ConfigFileName = "config.json"

// ConfigData represents struct of config file
type ConfigData struct {
	// Simple is S category of checks,
	// contains all checks that are concerned with simplifying code.
	Simple []string
	// Stylecheck is ST category of checks,
	// contains all checks that are concerned with stylistic issues
	Stylecheck []string
	// Quickfix is QF category of checks, codenamed quickfix,
	// contains checks that are used as part of gopls for automatic refactorings.
	Quickfix []string
}

// ConfigInit reads config in JSON format and parses it.
func ConfigInit() (*ConfigData, error) {
	data, err := os.ReadFile(ConfigFileName)
	if err != nil {
		return nil, err
	}

	var cfg ConfigData

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func main() {
	// initialize config
	cfg, err := ConfigInit()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// set required analyzers
	mychecks := []*analysis.Analyzer{
		// check for call os.Exit in main function
		osexit.Analyzer,
		// checks consistency of Printf format strings and arguments
		printf.Analyzer,
		// check for possible unintended shadowing of variables
		shadow.Analyzer,
		// checks struct field tags are well-formed
		structtag.Analyzer,
	}

	checks := make(map[string]bool)

	// init simple checks
	for _, v := range cfg.Simple {
		checks[v] = true
	}
	// init stylecheck checks
	for _, v := range cfg.Stylecheck {
		checks[v] = true
	}
	// init quickfix checks
	for _, v := range cfg.Quickfix {
		checks[v] = true
	}

	// add category SA* of staticcheck
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	// add some Simple checks
	for _, v := range simple.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	// add some Stylecheck checks
	for _, v := range stylecheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	// add some Quickfix checks
	for _, v := range quickfix.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
