package cmd

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Runs any of the checks, and prints violations, if any",
	RunE: runCheckCmd,
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

	// checks for default vars on modules
	err = checkForDefaultVars(tfFiles)
	if err != nil {
		return err
	}

	return nil
}

func checkForDefaultVars(files []string) error {
	referencedModules, err := getReferencedModules(files)
	if err != nil {
		return err
	}

	if len(referencedModules) == 0 {
		return nil
	}

	for _, mod := range referencedModules {
		err = checkForDefaultVarsInModule(mod)
		if err != nil {
			return err
		}

	}

	return nil
}

func checkForDefaultVarsInModule(module module) error {
	definedVars, err := getModuleVariables(module)
	if err != nil {
		return err
	}

	dvMap := toMap(definedVars)
	referencedVars := getAttributes(module)
	fmt.Println("The following referenced vars are the same as the default for module " + nameM(module))
	for attrName, dv := range referencedVars {
		if rv := dvMap[attrName]; rv != nil && equals(dv, rv) {
			fmt.Printf("\t%v\n", nameV(rv))
		}
	}

	return nil
}

func toMap(vars []variable) map[string]variable {
	m := make(map[string]variable)
	for _, v := range vars {
		m[nameV(v)] = v
	}

	return m
}

func equals(attr *hclwrite.Attribute, definedVar variable) bool {
	var dv *hclwrite.Block
	dv = definedVar

	// fmt.Println(dv.Labels()[0])
	definedDefaultValue := dv.Body().GetAttribute("default")
	if definedDefaultValue == nil {
		return false
	}

	// fmt.Println(attributeToString(attr))
	// fmt.Println(attributeToString(definedDefaultValue))

	return attributeToString(attr) == attributeToString(definedDefaultValue)

	// return true
}

func attributeToString(definedDefaultValue *hclwrite.Attribute) string {
	tokens := definedDefaultValue.Expr().BuildTokens(nil)
	return string(tokens.Bytes()[:])
}
