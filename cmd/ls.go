package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List discovered modules",
	RunE:  runLsCmd,
}

var resolveVariables bool

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().BoolVarP(&resolveVariables, "resolve-variables", "r", false, "Resolve variables for modules")
}

func runLsCmd(cmd *cobra.Command, args []string) error {
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

	if verbose {
		fmt.Println("Discovered TF files: ")
		for _, m := range matches {
			fmt.Printf("\t%v\n", m)
		}
	}

	referencedModules, err := getReferencedModules(matches)
	if err != nil {
		return err
	}

	if len(referencedModules) == 0 {
		fmt.Println("No modules detected in any of the TF files")
		return nil
	}

	fmt.Println("Detected modules:")
	for _, mod := range referencedModules {
		fmt.Printf("\t%v\n", mod)
	}

	if resolveVariables {
		fmt.Printf("\n\nVariables for modules:\n")

		for _, mod := range referencedModules {
			fmt.Printf("\n%v:\n", mod)
			vars, err := getModuleVariables(mod)
			if err != nil {
				return err
			}
			for _, v := range vars {
				fmt.Printf("\t%v\n", v)
			}
		}
	}

	return nil
}
