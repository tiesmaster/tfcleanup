package cmd

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type formatUsages map[string][]formatInvocation

type formatInvocation struct {
	expr   hcl.Expression
	tokens []hclsyntax.Token
}

// CHECK

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
			return tokens[start : i+1]
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

// FIX

func convertFormatUsageToInterpolation(report formatUsages) error {
	for filename, formatInvocations := range report {
		if len(formatInvocations) == 0 {
			continue
		}
		err := convertFormatUsageToInterpolationForFile(filename, formatInvocations)
		if err != nil {
			return err
		}
	}
	return nil
}

func convertFormatUsageToInterpolationForFile(filename string, formatInvocations []formatInvocation) error {
	return patchFile(filename, func(hclFile *hclwrite.File) (*hclwrite.File, error) {
		for _, fi := range formatInvocations {
			bl, attrName := getAttributeForWrite(hclFile, fi.expr)
			bl.Body().SetAttributeRaw(attrName, convertFormatToInterpolation(fi.tokens))
		}
		return hclFile, nil
	})
}

func getAttributeForWrite(hclFile *hclwrite.File, expr hcl.Expression) (hclwrite.Block, string) {
	panic("unimplemented")
}

func convertFormatToInterpolation(tokens []hclsyntax.Token) hclwrite.Tokens {
	var resultTokens []*hclwrite.Token
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		fmt.Printf("%s: %s\n", t.Type, string(t.Bytes))
		if t.Type == hclsyntax.TokenIdent && string(t.Bytes) == "format" {
			newTokens, consumedTokens := parseFormatAndReturnInterpolationTokens(tokens[i:])
			resultTokens = append(resultTokens, newTokens...)
			i += consumedTokens - 1 // account for the next loop increase
		} else {
			token := toHclwriteToken(t)
			resultTokens = append(resultTokens, &token)
		}
	}
	return resultTokens
}

func parseFormatAndReturnInterpolationTokens(tokens []hclsyntax.Token) ([]*hclwrite.Token, int) {
	i := 0

	// eat format token
	i++

	// eat open bracket token
	i++

	var newTokens []*hclwrite.Token
	var token hclwrite.Token

	// consume opening quote token
	token = toHclwriteToken(tokens[i])
	newTokens = append(newTokens, &token)
	i++;

	// consume quoted literal
	token = toHclwriteToken(tokens[i])
	newTokens = append(newTokens, &token)
	i++;

	// consume closing quote token
	token = toHclwriteToken(tokens[i])
	newTokens = append(newTokens, &token)
	i++;

	// eat close bracket token
	i++

	return newTokens, i

	///

	// take the main format string
	t := tokens[i]
	fmt.Println(t)

	// get all args
	for ; tokens[i].Type != hclsyntax.TokenOParen; i++ {

	}

	return nil, 0
}

func toHclwriteToken(token hclsyntax.Token) hclwrite.Token {
	return hclwrite.Token{
		Type:  token.Type,
		Bytes: token.Bytes,
	}
}
