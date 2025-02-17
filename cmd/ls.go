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

	referencedModules, err := getReferencedModules(matches)
	if err != nil {
		return err
	}

	if len(referencedModules) > 0 {
		fmt.Println("Detected modules:")
		for _, m := range referencedModules {
			fmt.Printf("\t%v\n", m)
		}
	} else {
		fmt.Println("No modules detected in any of the TF files")
	}

	return nil
}

func getReferencedModules(filenames []string) ([]string, error) {
	var allModules []string
	for _, f := range filenames {
		modules, err := getReferencedModulesForFile(f)
		if err != nil {
			return nil, err
		}
		allModules = append(allModules, modules...)
	}

	return allModules, nil
}

func getReferencedModulesForFile(filename string) ([]string, error) {
	input, _ := os.ReadFile(filename)
	hclFile, diags := hclwrite.ParseConfig(input, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, errors.New("failed to parse TF file: " + diags.Error())
	}

	hclBody := hclFile.Body()

	var modules []string
	for _, bl := range hclBody.Blocks() {
		if bl.Type() == "module" {
			modules = append(modules, bl.Labels()[0])
		}
	}

	return modules, nil
}
