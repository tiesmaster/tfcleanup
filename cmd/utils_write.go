package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func patchFile(filename string, patch func(hclFile *hclwrite.File) (*hclwrite.File, error)) error {
	input, _ := os.ReadFile(filename)
	hclFile, diags := hclwrite.ParseConfig(input, filename, hcl.Pos{Line: 1, Column: 1})

	if diags.HasErrors() {
		return errors.New("failed to parse TF file: " + diags.Error())
	}

	newHclFile, err := patch(hclFile)
	if err != nil {
		return err
	}

	if err = os.WriteFile(filename, newHclFile.Bytes(), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write file: %s", err)
	}

	return nil
}

func getModuleBlockForWrite(hclFile *hclwrite.File, mod module) *hclwrite.Block {
	for _, bl := range hclFile.Body().Blocks() {
		if bl.Type() == "module" && bl.Labels()[0] == mod.name() {
			return bl
		}
	}
	return nil
}
