package cmd

import (
	"errors"
	"io/fs"
	"os"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func ensureTargetDir() error {
	if targetDir != "" {
		err := os.Chdir(targetDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTerraformFiles() ([]string, error) {
	dir := os.DirFS(".")
	matches, err := fs.Glob(dir, "*.tf")
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, errors.New("no TF files detected")
	}

	return matches, nil
}

type module struct {
	bl *hclsyntax.Block
}

type variableDefinition struct {
	bl *hclsyntax.Block
}

func getReferencedModules(filenames []string) ([]module, error) {
	var allModules []module
	for _, f := range filenames {
		modules, err := readModules(f)
		if err != nil {
			return nil, err
		}
		allModules = append(allModules, modules...)
	}

	return allModules, nil
}

func readModules(filename string) ([]module, error) {
	moduleBlocks, err := getBlocksFromFile(filename, "module")
	if err != nil {
		return nil, err
	}

	var modules []module
	for _, bl := range moduleBlocks {
		modules = append(modules, module{bl})
	}

	return modules, nil
}

func getModuleVariables(mod module) ([]variableDefinition, error) {
	panic("not implemented")
}

func (mod module) name() string {
	return blockName(mod.bl)
}

func (v variableDefinition) name() string {
	return blockName(v.bl)
}

func blockName(bl *hclsyntax.Block) string {
	return bl.Labels[0]
}

func getBlocksFromFile(filename, blockName string) ([]*hclsyntax.Block, error) {
	hclFile, diags := hclparse.NewParser().ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, errors.New("failed to parse TF file: " + diags.Error())
	}

	hclBody := hclFile.Body.(*hclsyntax.Body)

	var blocks []*hclsyntax.Block
	for _, bl := range hclBody.Blocks {
		if bl.Type == blockName {
			blocks = append(blocks, bl)
		}
	}

	return blocks, nil
}
