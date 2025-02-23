package cmd

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type formatUsages map[string][]int

func checkForFormatUsage(files []string) (formatUsages, error) {
	affectedFiles := make(map[string][]int)
	for _, f := range files {
		affectedLines, err := checkForFormatUsageInFile(f)
		if err != nil {
			return nil, err
		}

		if len(affectedLines) > 0 {
			affectedFiles[f] = affectedLines
		}
	}
	return affectedFiles, nil
}

func checkForFormatUsageInFile(file string) ([]int, error) {
	tokens, err := readHclTokens(file)
	if err != nil {
		return nil, err
	}

	var affectedLines []int
	for _, token := range tokens {
		if isFormatToken(token) {
			affectedLines = append(affectedLines, token.Range.Start.Line)
		}
	}

	return affectedLines, nil
}

func isFormatToken(token hclsyntax.Token) bool {
	return token.Type == hclsyntax.TokenIdent && isTokenText(token, "format")
}
