package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
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

type module *hclwrite.Block

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
		modules = append(modules, bl)
	}

	return modules, nil
}

func getModuleVariables(mod module) ([]string, error) {
	moduleDir := path.Join(".terraform/modules/", name(mod))
	matches, err := fs.Glob(os.DirFS(moduleDir), "*.tf")
	if err != nil {
		return nil, err
	}

	var allVariables []string
	for _, m := range matches {
		vars, err := readVariables(path.Join(moduleDir, m))
		if err != nil {
			return nil, err
		}
		allVariables = append(allVariables, vars...)
	}

	return allVariables, nil
}

func name(mod module) string {
	var bl *hclwrite.Block
	bl = mod
	return bl.Labels()[0]
}

func readVariables(filename string) ([]string, error) {
	variableBlocks, err := getBlocksFromFile(filename, "variable")
	if err != nil {
		return nil, err
	}

	var variables []string
	for _, bl := range variableBlocks {
		variables = append(variables, bl.Labels()[0])
	}

	return variables, nil
}

func getBlocksFromFile(filename, blockName string) ([]*hclwrite.Block, error) {
	input, _ := os.ReadFile(filename)
	hclFile, diags := hclwrite.ParseConfig(input, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, errors.New("failed to parse TF file: " + diags.Error())
	}

	hclBody := hclFile.Body()

	var blocks []*hclwrite.Block
	for _, bl := range hclBody.Blocks() {
		if bl.Type() == blockName {
			blocks = append(blocks, bl)
		}
	}

	return blocks, nil
}
