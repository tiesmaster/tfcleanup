package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tfcleanup",
	Short: "Simple CLI to clean some obvious things in terraform files",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var targetDir string
var verbose bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&targetDir, "target-dir", "t", "", "target dir (default is current working directory)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print verbose output")
}
