package gofmt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"io/ioutil"
	"os"
	"strings"
)

func Run(filename string, needSimplify bool) ([]byte, error) {
	src, err := ioutil.ReadFile(filename)
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

func RunRewrite(filename string, needSimplify bool, rewrite string) ([]byte, error) {
	rewriter, err := getRewrite(rewrite)
	if err != nil {
		return nil, err
	}

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

	if rewriter != nil {
		file = rewriter(file)
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

func getRewrite(rewriteRule string) (func(*ast.File) *ast.File, error) {
	if rewriteRule == "" {
		return nil, nil
	}

	f := strings.Split(rewriteRule, "->")
	if len(f) != 2 {
		return nil, fmt.Errorf("rewrite rule must be of the form 'pattern -> replacement'\n")
	}
	pattern, err := parseExprErr(f[0], "pattern")
	if err != nil {
		return nil, err
	}

	replace, err := parseExprErr(f[1], "replacement")
	if err != nil {
		return nil, err
	}

	return func(p *ast.File) *ast.File { return rewriteFile(pattern, replace, p) }, nil
}

func parseExprErr(s, what string) (ast.Expr, error) {
	x, err := parser.ParseExpr(s)
	if err != nil {
		return nil, fmt.Errorf("parsing %s %s at %s\n", what, s, err)
	}
	return x, nil
}
