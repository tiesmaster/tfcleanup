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

type moduleLegacy struct {
	bl       *hclwrite.Block
	filename string
}

type variableDefinitionLegacy struct {
	bl *hclwrite.Block
}

type expressionLegacy struct {
	attr *hclwrite.Attribute
}

func getReferencedModulesLegacy(filenames []string) ([]moduleLegacy, error) {
	var allModules []moduleLegacy
	for _, f := range filenames {
		modules, err := readModulesLegacy(f)
		if err != nil {
			return nil, err
		}
		allModules = append(allModules, modules...)
	}

	return allModules, nil
}

func readModulesLegacy(filename string) ([]moduleLegacy, error) {
	moduleBlocks, err := getBlocksFromFileLegacy(filename, "module")
	if err != nil {
		return nil, err
	}

	var modules []moduleLegacy
	for _, bl := range moduleBlocks {
		modules = append(modules, moduleLegacy{bl, filename})
	}

	return modules, nil
}

func getModuleVariablesLegacy(mod moduleLegacy) ([]variableDefinitionLegacy, error) {
	moduleDir := path.Join(".terraform/modules/", mod.name())
	matches, err := fs.Glob(os.DirFS(moduleDir), "*.tf")
	if err != nil {
		return nil, err
	}

	var allVariables []variableDefinitionLegacy
	for _, m := range matches {
		vars, err := readVariablesLegacy(path.Join(moduleDir, m))
		if err != nil {
			return nil, err
		}
		allVariables = append(allVariables, vars...)
	}

	return allVariables, nil
}

func (mod moduleLegacy) name() string {
	return blockNameLegacy(mod.bl)
}

func (v variableDefinitionLegacy) name() string {
	return blockNameLegacy(v.bl)
}

func blockNameLegacy(bl *hclwrite.Block) string {
	return bl.Labels()[0]
}

func (v variableDefinitionLegacy) defaultValue() *expressionLegacy {
	defaultValue := v.bl.Body().GetAttribute("default")
	if defaultValue == nil {
		return nil
	}

	return &expressionLegacy{defaultValue}
}

func readVariablesLegacy(filename string) ([]variableDefinitionLegacy, error) {
	variableBlocks, err := getBlocksFromFileLegacy(filename, "variable")
	if err != nil {
		return nil, err
	}

	var variables []variableDefinitionLegacy
	for _, bl := range variableBlocks {
		variables = append(variables, variableDefinitionLegacy{bl})
	}

	return variables, nil
}

func getBlocksFromFileLegacy(filename, blockName string) ([]*hclwrite.Block, error) {
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

func readHclFileLegacy(filename string) (*hclwrite.File, error) {
	input, _ := os.ReadFile(filename)
	hclFile, diags := hclwrite.ParseConfig(input, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, errors.New("failed to parse TF file: " + diags.Error())
	}

	return hclFile, nil
}

func getVariableAssignmentsLegacy(module moduleLegacy) map[string]expressionLegacy {
	m := module.bl.Body().Attributes()

	newMap := make(map[string]expressionLegacy)
	for k, v := range m {
		newMap[k] = expressionLegacy{v}
	}

	return newMap
}

func equalToVariableDefinitionLegacy(assignExpr expressionLegacy, varDefinition variableDefinitionLegacy) bool {
	defaultValue := varDefinition.defaultValue()

	if defaultValue == nil {
		return false
	}

	return assignExpr.exprToString() == defaultValue.exprToString()
}

func (expr expressionLegacy) exprToString() string {
	tokens := expr.attr.Expr().BuildTokens(nil)
	return string(tokens.Bytes())
}

func patchFileLegacy(filename string, patch func(hclFile *hclwrite.File) (*hclwrite.File, error)) error {
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

func getModuleBlockLegacy(hclFile *hclwrite.File, mod moduleLegacy) *hclwrite.Block {
	for _, bl := range hclFile.Body().Blocks() {
		if bl.Type() == "module" && bl.Labels()[0] == mod.name() {
			return bl
		}
	}
	return nil
}

func isTokenTextLegacy(token *hclwrite.Token, text string) bool {
	return string(token.Bytes) == text
}
