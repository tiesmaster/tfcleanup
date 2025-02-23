package cmd

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type unneededAttrAssigs map[moduleLegacy][]string

func checkForUnneededAttributeAssignments(files []string) (unneededAttrAssigs, error) {
	referencedModules, err := getReferencedModulesLegacy(files)
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

		m[mod] = getAttrNames(unneededAssignments)
	}

	return m, nil
}

func getAttrNames(unneededAssignments []variableDefinitionLegacy) []string {
	var names []string
	for _, assign := range unneededAssignments {
		names = append(names, assign.name())
	}
	return names
}

func checkForUnneededAssignments(module moduleLegacy) ([]variableDefinitionLegacy, error) {
	moduleVariables, err := getModuleVariablesLegacy(module)
	if err != nil {
		return nil, err
	}

	moduleVariablesMap := toMap(moduleVariables)
	variableAssignments := getVariableAssignmentsLegacy(module)

	variableAssignments = filterForTerraformAssignments(variableAssignments)


	var unneededAssignments []variableDefinitionLegacy
	for varName, assignExpr := range variableAssignments {
		if varDefinition, exists := moduleVariablesMap[varName]; exists && equalToVariableDefinitionLegacy(assignExpr, varDefinition) {
			unneededAssignments = append(unneededAssignments, varDefinition)
		} else if !exists && verbose {
			fmt.Printf("WARNING: module assignment not found as variable in referenced module '%v': %v\n", module.name(), varName)
		}
	}

	return unneededAssignments, nil
}

func filterForTerraformAssignments(variableAssignments map[string]expressionLegacy) map[string]expressionLegacy {
	delete(variableAssignments, "source")
	delete(variableAssignments, "version")

	return variableAssignments
}

func toMap(vars []variableDefinitionLegacy) map[string]variableDefinitionLegacy {
	m := make(map[string]variableDefinitionLegacy)
	for _, v := range vars {
		m[v.name()] = v
	}

	return m
}

func removeUnneededAttributes(report unneededAttrAssigs) error {
	for mod, unneededVars := range report {
		if len(unneededVars) == 0 {
			continue
		}
		err := removeUnneededAttributesFromModule(mod, unneededVars)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeUnneededAttributesFromModule(mod moduleLegacy, unneededVars []string) error {
	return patchFileLegacy(mod.filename, func(hclFile *hclwrite.File) (*hclwrite.File, error) {
		moduleBlock := getModuleBlockLegacy(hclFile, mod)
		for _, v := range unneededVars {
			moduleBlock.Body().RemoveAttribute(v)
		}

		return hclFile, nil
	})
}
