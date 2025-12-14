package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"sync"
)

type Options struct {
	NeedSimplify bool
	RewriteRules []RewriteRule
}

var parserModeMu sync.RWMutex

type RewriteRule struct {
	Pattern     string
	Replacement string
}

// Source formats the code like gofmt.
// Empty string `rewrite` will be ignored.
func Source(filename string, src []byte, opts Options) ([]byte, error) {
	fset := token.NewFileSet()

	parserModeMu.Lock()
	initParserMode()
	parserModeMu.Unlock()

	file, sourceAdj, indentAdj, err := parse(fset, filename, src, false)
	if err != nil {
		return nil, err
	}

	file, err = rewriteFileContent(fset, file, opts.RewriteRules)
	if err != nil {
		return nil, err
	}

	ast.SortImports(fset, file)

	if opts.NeedSimplify {
		simplify(file)
	}

	return format(fset, file, sourceAdj, indentAdj, src, printer.Config{Mode: printerMode, Tabwidth: tabWidth})
}

func rewriteFileContent(fset *token.FileSet, file *ast.File, rewriteRules []RewriteRule) (*ast.File, error) {
	for _, rewriteRule := range rewriteRules {
		pattern, err := parseExpression(rewriteRule.Pattern, "pattern")
		if err != nil {
			return nil, err
		}

		replacement, err := parseExpression(rewriteRule.Replacement, "replacement")
		if err != nil {
			return nil, err
		}

		file = rewriteFile(fset, pattern, replacement, file)
	}

	return file, nil
}

func parseExpression(s, what string) (ast.Expr, error) {
	x, err := parser.ParseExpr(s)
	if err != nil {
		return nil, fmt.Errorf("parsing %s %q at %s\n", what, s, err)
	}
	return x, nil
}
