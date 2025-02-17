package cmd

import (
	"fmt"

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

func checkForDefaultVarsInModule(module string) error {
	moduleVars, err := getModuleVariables(module)
	if err != nil {
		return err
	}

	fmt.Println(moduleVars)

	return nil
}
