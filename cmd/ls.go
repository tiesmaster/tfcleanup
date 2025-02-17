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

func init() {
	rootCmd.AddCommand(lsCmd)
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

	fmt.Println("Discovered TF files: ")
	fmt.Println(matches)

	return nil
}
