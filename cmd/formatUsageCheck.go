package cmd

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func checkForFormatUsage(files []string) ([]string, error) {
	var affectedFiles []string
	for _, f := range files {
		affected, err := checkForFormatUsageInFile(f)
		if err != nil {
			return nil, err
		}

		if affected {
			affectedFiles = append(affectedFiles, f)
		}
	}
	return affectedFiles, nil
}

func checkForFormatUsageInFile(file string) (bool, error) {
	hclFile, err := readHclFile(file)
	if err != nil {
		return false, err
	}

	tokens := hclFile.BuildTokens(nil)
	for _, token := range tokens {
		if isFormatToken(token) {
			return true, nil
		}
	}

	return false, nil
}

func isFormatToken(token *hclwrite.Token) bool {
	return token.Type == hclsyntax.TokenIdent && isTokenText(token, "format")
}
