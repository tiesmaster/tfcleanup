package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

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
	moduleDir := path.Join(".terraform/modules/", mod.name())
	matches, err := fs.Glob(os.DirFS(moduleDir), "*.tf")
	if err != nil {
		return nil, err
	}

	var allVariables []variableDefinition
	for _, m := range matches {
		vars, err := readVariables(path.Join(moduleDir, m))
		if err != nil {
			return nil, err
		}
		allVariables = append(allVariables, vars...)
	}

	return allVariables, nil
}

func (mod module) name() string {
	return blockName(mod.bl)
}

func (v variableDefinition) name() string {
	return blockName(v.bl)
}

func (mod module) location() string {
	return fmt.Sprintf("%v:%v", mod.bl.Range().Filename, mod.bl.Range().Start.Line)
}

func blockName(bl *hclsyntax.Block) string {
	return bl.Labels[0]
}

func readVariables(filename string) ([]variableDefinition, error) {
	variableBlocks, err := getBlocksFromFile(filename, "variable")
	if err != nil {
		return nil, err
	}

	var variables []variableDefinition
	for _, bl := range variableBlocks {
		variables = append(variables, variableDefinition{bl})
	}

	return variables, nil
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
