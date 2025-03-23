package cmd

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type formatUsages map[string][]formatInvocation

type formatInvocation struct {
	expr   hcl.Expression
	tokens []hclsyntax.Token
}

func checkForFormatUsage(files []string) (formatUsages, error) {
	result := make(map[string][]formatInvocation)
	for _, f := range files {
		violations, err := checkForFormatUsageInFile(f)
		if err != nil {
			return nil, err
		}

		if len(violations) > 0 {
			result[f] = violations
		}
	}
	return result, nil
}

func checkForFormatUsageInFile(file string) ([]formatInvocation, error) {
	tokens, err := readHclTokens(file)
	if err != nil {
		return nil, err
	}

	var violationTokens []hclsyntax.Token
	for _, token := range tokens {
		if isFormatToken(token) {
			violationTokens = append(violationTokens, token)
		}
	}

	result, err := convertTokensToExpressions(file, tokens, violationTokens)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func convertTokensToExpressions(filename string, tokens []hclsyntax.Token, violationTokens []hclsyntax.Token) ([]formatInvocation, error) {
	hclFile, diags := hclparse.NewParser().ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, errors.New("failed to parse TF file: " + diags.Error())
	}

	var result []formatInvocation
	for _, t := range violationTokens {
		expr := hclFile.OutermostExprAtPos(t.Range.Start)

		result = append(result, formatInvocation{expr, getAllTokensForExpression(tokens, expr)})
	}
	return result, nil
}

func getAllTokensForExpression(tokens []hclsyntax.Token, expr hcl.Expression) []hclsyntax.Token {
	start := 0
	for i, t := range tokens {
		if t.Range.Start == expr.Range().Start {
			start = i
		}

		if t.Range.End == expr.Range().End {
			return tokens[start:i+1]
		}
	}

	panic("cannot reach")
}

func isFormatToken(token hclsyntax.Token) bool {
	return token.Type == hclsyntax.TokenIdent && isTokenText(token, "format")
}

func (invoke formatInvocation) string() string {
	var str string
	for _, s := range invoke.tokens {
		str = str + string(s.Bytes)
	}
	return str
}

func (invoke formatInvocation) location() string {
	r := invoke.expr.Range()
	return fmt.Sprintf("%v:L%d:%d-%d", r.Filename, r.Start.Line, r.Start.Column, r.End.Column)
}
