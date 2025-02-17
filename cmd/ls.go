package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
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

	if verbose {
		fmt.Println("Discovered TF files: ")
		for _, m := range matches {
			fmt.Printf("\t%v\n", m)
		}
	}

	referencedModules := getReferencedModules(matches)
	fmt.Println("Detected modules:")
	for _, m := range referencedModules {
		fmt.Printf("\t%v\n", m)
	}

	return nil
}

func getReferencedModules(filenames []string) []string {
	var modules []string
	for _, f := range filenames {
		modules = append(modules, getReferencedModulesForFile(f)...)
	}

	return modules
}

func getReferencedModulesForFile(filename string) []string {
	input, _ := os.ReadFile(filename)
	hclFile, _ := hclwrite.ParseConfig(input, filename, hcl.Pos{Line: 1, Column: 1})

	hclBody := hclFile.Body()

	var modules []string
	for _, bl := range hclBody.Blocks() {
		if bl.Type() == "module" {
			modules = append(modules, bl.Labels()[0])
		}
	}

	return modules
}
