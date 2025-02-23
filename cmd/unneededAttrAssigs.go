package cmd

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type unneededAttrAssigs map[module][]expression

func checkForUnneededAttributeAssignments(files []string) (unneededAttrAssigs, error) {
	referencedModules, err := getReferencedModules(files)
	if err != nil {
		return nil, err
	}

	m := make(unneededAttrAssigs)
	if len(referencedModules) == 0 {
		return m, nil
	}

	for _, mod := range referencedModules {
		unneededAssignments, err := checkForUnneededAssignments(mod)
		if err != nil {
			return nil, err
		}

		m[mod] = unneededAssignments
	}

	return m, nil
}

func checkForUnneededAssignments(module module) ([]expression, error) {
	moduleVariables, err := getModuleVariables(module)
	if err != nil {
		return nil, err
	}

	moduleVariablesMap := toMap(moduleVariables)
	variableAssignments := getVariableAssignments(module)

	variableAssignments = filterForTerraformAssignments(variableAssignments)


	var unneededAssignments []expression
	for varName, assignExpr := range variableAssignments {
		if varDefinition, exists := moduleVariablesMap[varName]; exists && equalToVariableDefinition(assignExpr, varDefinition) {
			unneededAssignments = append(unneededAssignments, assignExpr)
		} else if !exists && verbose {
			fmt.Printf("WARNING: module assignment not found as variable in referenced module '%v': %v\n", module.name(), varName)
		}
	}

	return unneededAssignments, nil
}

func filterForTerraformAssignments(variableAssignments map[string]expression) map[string]expression {
	delete(variableAssignments, "source")
	delete(variableAssignments, "version")

	return variableAssignments
}

func toMap(vars []variableDefinition) map[string]variableDefinition {
	m := make(map[string]variableDefinition)
	for _, v := range vars {
		m[v.name()] = v
	}

	return m
}

func removeUnneededAttributes(report unneededAttrAssigs) error {
	for mod, unneededAssign := range report {
		if len(unneededAssign) == 0 {
			continue
		}
		err := removeUnneededAttributesFromModule(mod, unneededAssign)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeUnneededAttributesFromModule(mod module, unneededAssigns []expression) error {
	return patchFile(mod.filename(), func(hclFile *hclwrite.File) (*hclwrite.File, error) {
		moduleBlock := getModuleBlockForWrite(hclFile, mod)
		for _, assign := range unneededAssigns {
			moduleBlock.Body().RemoveAttribute(assign.name())
		}

		return hclFile, nil
	})
}
