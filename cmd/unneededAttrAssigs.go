package cmd

import "fmt"

type unneededAttrAssigs map[module][]string

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

		m[mod] = getAttrNames(unneededAssignments)
	}

	return m, nil
}

func getAttrNames(unneededAssignments []variableDefinition) []string {
	var names []string
	for _, assign := range unneededAssignments {
		names = append(names, assign.name())
	}
	return names
}

func checkForUnneededAssignments(module module) ([]variableDefinition, error) {
	moduleVariables, err := getModuleVariables(module)
	if err != nil {
		return nil, err
	}

	moduleVariablesMap := toMap(moduleVariables)
	variableAssignments := getVariableAssignments(module)

	variableAssignments = filterForTerraformAssignments(variableAssignments)


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
