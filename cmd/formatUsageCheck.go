package cmd

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type formatUsages map[string][]formatInvocation

type formatInvocation struct {
	token hclsyntax.Token
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

	var violations []formatInvocation
	for _, token := range tokens {
		if isFormatToken(token) {
			violations = append(violations, formatInvocation{token}) // token.Range.Start.Line
		}
	}

	return violations, nil
}

func isFormatToken(token hclsyntax.Token) bool {
	return token.Type == hclsyntax.TokenIdent && isTokenText(token, "format")
}

func (invoke formatInvocation) string() string {
	return string(invoke.token.Bytes)
}

func (invoke formatInvocation) location() string {
	return fmt.Sprintf("%v:%d", invoke.token.Range.Filename, invoke.token.Range.Start.Line)
}
