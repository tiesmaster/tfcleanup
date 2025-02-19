package cmd

import (
	"errors"
	"fmt"
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

type module struct {
	bl       *hclwrite.Block
	filename string
}

type variableDefinition struct {
	bl *hclwrite.Block
}

type expression struct {
	attr *hclwrite.Attribute
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
		modules = append(modules, module{bl, filename})
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

func blockName(bl *hclwrite.Block) string {
	return bl.Labels()[0]
}

func (v variableDefinition) defaultValue() *expression {
	defaultValue := v.bl.Body().GetAttribute("default")
	if defaultValue == nil {
		return nil
	}

	return &expression{defaultValue}
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

func readHclFile(filename string) (*hclwrite.File, error) {
	input, _ := os.ReadFile(filename)
	hclFile, diags := hclwrite.ParseConfig(input, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, errors.New("failed to parse TF file: " + diags.Error())
	}

	return hclFile, nil
}

func getVariableAssignments(module module) map[string]expression {
	m := module.bl.Body().Attributes()

	newMap := make(map[string]expression)
	for k, v := range m {
		newMap[k] = expression{v}
	}

	return newMap
}

func equalToVariableDefinition(assignExpr expression, varDefinition variableDefinition) bool {
	defaultValue := varDefinition.defaultValue()

	if defaultValue == nil {
		return false
	}

	return assignExpr.exprToString() == defaultValue.exprToString()
}

func (expr expression) exprToString() string {
	tokens := expr.attr.Expr().BuildTokens(nil)
	return string(tokens.Bytes())
}

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

func getModuleBlock(hclFile *hclwrite.File, mod module) *hclwrite.Block {
	for _, bl := range hclFile.Body().Blocks() {
		if bl.Type() == "module" && bl.Labels()[0] == mod.name() {
			return bl
		}
	}
	return nil
}

func isTokenText(token *hclwrite.Token, text string) bool {
	return string(token.Bytes) == text
}
