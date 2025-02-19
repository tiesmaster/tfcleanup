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
	report, err := checkForUnneededAttributeAssignments(tfFiles)
	if err != nil {
		return err
	}

	printReport(report)

	return nil
}

func printReport(report unneededAttrAssigs) {
	if len(report) == 0 {
		fmt.Println("No unneeded module assignments were found")
		return
	}

	for mod, unneededVars := range report {
		if len(unneededVars) == 0 {
			fmt.Printf("\nNo unneeded module assignments were found for module '%v'\n", mod.name())
			continue
		}

		fmt.Printf("\nThe following module assignments are unneeded for module '%v':\n", mod.name())

		for _, v := range unneededVars {
			fmt.Printf("\t%v\n", v.name())
		}
	}
}
