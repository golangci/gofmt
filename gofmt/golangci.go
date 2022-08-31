package gofmt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

// Run runs gofmt.
// Deprecated: use RunRewrite instead.
func Run(filename string, needSimplify bool) ([]byte, error) {
	return RunRewrite(filename, needSimplify, "")
}

// RunRewrite runs gofmt.
// empty string `rewrite` will be ignored.
func RunRewrite(filename string, needSimplify bool, rewriteRule string) ([]byte, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()

	initParserMode()

	file, sourceAdj, indentAdj, err := parse(fset, filename, src, false)
	if err != nil {
		return nil, err
	}

	file, err = rewriteFileContent(fset, rewriteRule, file)
	if err != nil {
		return nil, err
	}

	ast.SortImports(fset, file)

	if needSimplify {
		simplify(file)
	}

	res, err := format(fset, file, sourceAdj, indentAdj, src, printer.Config{Mode: printerMode, Tabwidth: tabWidth})
	if err != nil {
		return nil, err
	}

	if bytes.Equal(src, res) {
		return nil, nil
	}

	// formatting has changed
	data, err := diffWithReplaceTempFile(src, res, filename)
	if err != nil {
		return nil, fmt.Errorf("error computing diff: %s", err)
	}

	return data, nil
}

func rewriteFileContent(fset *token.FileSet, rewriteRule string, file *ast.File) (*ast.File, error) {
	if rewriteRule == "" {
		return file, nil
	}

	f := strings.Split(rewriteRule, "->")
	if len(f) != 2 {
		return nil, fmt.Errorf("rewrite rule must be of the form 'pattern -> replacement'\n")
	}

	pattern, err := parseExpression(f[0], "pattern")
	if err != nil {
		return nil, err
	}

	replace, err := parseExpression(f[1], "replacement")
	if err != nil {
		return nil, err
	}

	return rewriteFile(fset, pattern, replace, file), nil
}

func parseExpression(s, what string) (ast.Expr, error) {
	x, err := parser.ParseExpr(s)
	if err != nil {
		return nil, fmt.Errorf("parsing %s %s at %s\n", what, s, err)
	}
	return x, nil
}
