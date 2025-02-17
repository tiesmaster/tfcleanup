package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

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
	if targetDir == "" {
		return fmt.Errorf("target dir cannot be empty (yet)")
	}

	err := os.Chdir(targetDir)
	if err != nil {
		return err
	}

	dir := os.DirFS(".")
	matches, _ := fs.Glob(dir, "*.tf")

	if len(matches) == 0 {
		return errors.New("no TF files detected")
	}

	// checks for default vars on modules
	err = checkForDefaultVars(matches)
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
