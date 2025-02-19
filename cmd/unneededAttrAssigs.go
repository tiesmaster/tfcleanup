package cmd

import "fmt"

type unneededAttrAssigs map[module][]variableDefinition

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

func checkForUnneededAssignments(module module) ([]variableDefinition, error) {
	moduleVariables, err := getModuleVariables(module)
	if err != nil {
		return nil, err
	}

	moduleVariablesMap := toMap(moduleVariables)
	variableAssignments := getVariableAssignments(module)

	// TODO: Filter for "well-known" attribute assignments that are TF specific (version, and source)

	var unneededAssignments []variableDefinition
	for varName, assignExpr := range variableAssignments {
		if varDefinition, exists := moduleVariablesMap[varName]; exists && equalToVariableDefinition(assignExpr, varDefinition) {
			unneededAssignments = append(unneededAssignments, varDefinition)
		} else if !exists && verbose {
			fmt.Printf("WARNING: module assignment not found as variable in referenced module '%v': %v\n", module.name(), varName)
		}
	}

	return unneededAssignments, nil
}

func toMap(vars []variableDefinition) map[string]variableDefinition {
	m := make(map[string]variableDefinition)
	for _, v := range vars {
		m[v.name()] = v
	}

	return m
}
