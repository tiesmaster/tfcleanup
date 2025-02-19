package cmd

import (
	"github.com/spf13/cobra"
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Applies fixes, if any violates are found",
	RunE:  runFixCmd,
}

func init() {
	rootCmd.AddCommand(fixCmd)
}

func runFixCmd(cmd *cobra.Command, args []string) error {
	err := ensureTargetDir()
	if err != nil {
		return err
	}

	tfFiles, err := getTerraformFiles()
	if err != nil {
		return err
	}

	report, err := checkForUnneededAttributeAssignments(tfFiles)
	if err != nil {
		return err
	}

	err = removeUnneededAttributes(report)
	if err != nil {
		return err
	}

	return nil
}
