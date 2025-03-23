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

	err = checkUnneededAttributeAssignments(tfFiles)
	if err != nil {
		return err
	}

	err = checkFormatUsages(tfFiles)
	if err != nil {
		return err
	}

	return nil
}

// checks for attribute assignments with default values of variable from their module
func checkUnneededAttributeAssignments(tfFiles []string) error {
	report, err := checkForUnneededAttributeAssignments(tfFiles)
	if err != nil {
		return err
	}

	if len(report) == 0 {
		fmt.Println("No unneeded module assignments were found")
		return nil
	}

	fmt.Println("== RESULTS FOR THE UNNEEDED MODULE ASSIGNMENTS ==")

	for mod, unneededAssigns := range report {
		if len(unneededAssigns) == 0 {
			fmt.Printf("\n\tNo unneeded module assignments were found for module '%v'\n", mod.name())
			continue
		}

		fmt.Printf("\n\tThe following module assignments are unneeded for module '%v':\n", mod.name())

		for _, assign := range unneededAssigns {
			fmt.Printf("\t\t%v", assign.name())
			if verbose {
				fmt.Printf(" (%v)", assign.location())
			}
			fmt.Println()
		}
	}

	return nil
}

// check for format() usage
func checkFormatUsages(tfFiles []string) error {
	report, err := checkForFormatUsage(tfFiles)
	if err != nil {
		return err
	}

	if len(report) == 0 {
		fmt.Println("No format() usages were found")
		return nil
	}

	fmt.Println("== RESULTS FOR FORMAT() USAGES ==")

	for filename, formatInvocations := range report {
		if len(formatInvocations) == 0 {
			fmt.Printf("\n\tNo format() usages were found for file '%v'\n", filename)
			continue
		}

		fmt.Printf("\n\tThe following format() usages were found for file '%v':\n", filename)

		for _, invoke := range formatInvocations {
			fmt.Printf("\t\t%v", invoke.string())
			if verbose {
				fmt.Printf(" (%v)", invoke.location())
			}
			fmt.Println()
		}
	}

	return nil
}
