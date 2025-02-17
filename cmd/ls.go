package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List discovered modules",
	RunE: runLsCmd,
}

func init() {
	rootCmd.AddCommand(lsCmd)
}

func runLsCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("ls called")
	fmt.Println("target dir: " + targetDir)

	return nil
}
