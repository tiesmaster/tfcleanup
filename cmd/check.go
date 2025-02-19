package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Runs any of the checks, and prints violations, if any",
	RunE:  runCheckCmd,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheckCmd(cmd *cobra.Command, args []string) error {
	err := ensureTargetDir()
	if err != nil {
		return err
	}

	tfFiles, err := getTerraformFiles()
	if err != nil {
		return err
	}

	// checks for attribute assignments with default values of variable from their module
	err = checkForUnneededAttributeAssignments(tfFiles)
	if err != nil {
		return err
	}

	return nil
}

func checkForUnneededAttributeAssignments(files []string) error {
	referencedModules, err := getReferencedModules(files)
	if err != nil {
		return err
	}

	if len(referencedModules) == 0 {
		return nil
	}

	for _, mod := range referencedModules {
		err = checkForUnneededAssignments(mod)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkForUnneededAssignments(module module) error {
	moduleVariables, err := getModuleVariables(module)
	if err != nil {
		return err
	}

	moduleVariablesMap := toMap(moduleVariables)
	variableAssignments := getVariableAssignments(module)

	// TODO: Filter for "well-known" attribute assignments that are TF specific (version, and source)

	fmt.Println("The following referenced vars are the same as the default for module " + nameM(module))
	for varName, assignExpr := range variableAssignments {
		if varDefinition, exists := moduleVariablesMap[varName]; exists && equalToVariableDefinition(assignExpr, varDefinition) {
			fmt.Printf("\t%v\n", nameV(varDefinition))
		} else if !exists && verbose {
			fmt.Printf("WARNING: module assignment not found as variable in referenced module: %v\n", varName)
		}
	}

	return nil
}

func toMap(vars []variableDefinition) map[string]variableDefinition {
	m := make(map[string]variableDefinition)
	for _, v := range vars {
		m[nameV(v)] = v
	}

	return m
}
