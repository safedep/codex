/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/safedep/codex/pkg/parser/py/imports"
	"github.com/safedep/dry/log"
	"github.com/safedep/vet/pkg/common/logger"
	"github.com/spf13/cobra"
)

var input_file string

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan your project or file to find direct dependencies",
	Long:  `Scan your project or file to find direct dependencies. It has few subdommands.`,
	Run: func(cmd *cobra.Command, args []string) {
		// findDirectDeps()
		log.Debugf("Running Scan cmd..")
	},
}

// scanCmd represents the scan command
var cmdScanFile = &cobra.Command{
	Use:   "file",
	Short: "Find direct dependencies of a single python file",
	Long:  `Find direct dependencies of a single python file`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Running Scan File..")
		scanFile()
	},
}

// scanCmd represents the scan command
var cmdDirectDeps = &cobra.Command{
	Use:   "find-direct-deps",
	Short: "Find direct dependencies of the project based on imported modules, not just package file",
	Long: `Find direct dependencies of the project based on imported modules, not just package file. 
	For example:
	go run main.go scan find-direct-deps --input <project_path>

	Imported Modules:
	json
	argparse
	feedparser
	mock
	shlex
	unittest
	git
	titlecase
	nc
	configparser

	Exported Modules:
	my_project


`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Running Direct Deps..")
		findDirectDeps()
	},
}

func init() {
	log.InitZapLogger("Zap")
	rootCmd.AddCommand(scanCmd)

	scanCmd.PersistentFlags().StringVar(&input_file, "input", "", "Provide  Github Acc Name")
	scanCmd.MarkPersistentFlagRequired("input")

	scanCmd.AddCommand(cmdDirectDeps)
	scanCmd.AddCommand(cmdScanFile)

}

func findDirectDeps() {
	filename := input_file

	ctx := context.Background()
	cf := imports.NewPyCodeParserFactory()
	parser, err := cf.NewCodeParser()
	if err != nil {
		logger.Warnf("Error while creating parser %v", err)
		return
	}
	includeExtensions := []string{".py"}
	excludeDirs := []string{".git", "test"}

	rootPkgs, _ := parser.FindImportedModules(ctx, filename, true, includeExtensions, excludeDirs)
	fmt.Println("Imported Modules:")
	for _, k := range rootPkgs.GetPackagesNames() {
		fmt.Println(k)
	}

	exportedModules, _ := parser.FindExportedModules(ctx, filename)
	fmt.Println("Exported Modules:")
	for _, k := range exportedModules.GetExportedModules() {
		fmt.Println(k)
	}
}

func scanFile() {
	ctx := context.Background()
	cf := imports.NewPyCodeParserFactory()
	parser, err := cf.NewCodeParser()
	if err != nil {
		logger.Warnf("Error while creating parser %v", err)
		return
	}

	basePath, filename := path.Split(input_file)
	parsedCode, err := parser.ParseFile(ctx, basePath, filename)
	if err != nil {
		panic(err)
	}

	parsedCode.ExtractModules()
	parsedCode.MakeMethodMap()
}
