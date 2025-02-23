package cmd

import (
	"errors"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func readHclFileLegacy(filename string) (*hclwrite.File, error) {
	input, _ := os.ReadFile(filename)
	hclFile, diags := hclwrite.ParseConfig(input, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, errors.New("failed to parse TF file: " + diags.Error())
	}

	return hclFile, nil
}

func isTokenTextLegacy(token *hclwrite.Token, text string) bool {
	return string(token.Bytes) == text
}
