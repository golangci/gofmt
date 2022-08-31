package gofmt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"os"
	"strings"
)

// Run runs gofmt.
// Deprecated: use RunRewrite instead.
func Run(filename string, needSimplify bool) ([]byte, error) {
	return RunRewrite(filename, needSimplify, nil)
}

// RunRewrite runs gofmt.
func RunRewrite(filename string, needSimplify bool, rewriteRules []string) ([]byte, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	initParserMode()

	file, sourceAdj, indentAdj, err := parse(fileSet, filename, src, false)
	if err != nil {
		return nil, err
	}

	ast.SortImports(fileSet, file)

	if needSimplify {
		simplify(file)
	}

	for _, rule := range rewriteRules {
		file, err = rewriteFileContent(rule, file)
		if err != nil {
			return nil, err
		}
	}

	res, err := format(fileSet, file, sourceAdj, indentAdj, src, printer.Config{Mode: printerMode, Tabwidth: tabWidth})
	if err != nil {
		return nil, err
	}

	if bytes.Equal(src, res) {
		return nil, nil
	}

	// formatting has changed
	data, err := diff(src, res, filename)
	if err != nil {
		return nil, fmt.Errorf("error computing diff: %s", err)
	}

	return data, nil
}

func rewriteFileContent(rewriteRule string, file *ast.File) (*ast.File, error) {
	f := strings.Split(rewriteRule, "->")
	if len(f) != 2 {
		return nil, fmt.Errorf("rewrite rule must be of the form 'pattern -> replacement', got %q instead", rewriteRule)
	}

	pattern, err := parseExpression(f[0], "pattern")
	if err != nil {
		return nil, err
	}

	replace, err := parseExpression(f[1], "replacement")
	if err != nil {
		return nil, err
	}

	return rewriteFile(pattern, replace, file), nil
}

func parseExpression(s, what string) (ast.Expr, error) {
	x, err := parser.ParseExpr(s)
	if err != nil {
		return nil, fmt.Errorf("parsing %s %s at %s\n", what, s, err)
	}
	return x, nil
}
