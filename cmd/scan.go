/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/safedep/codex/pkg/parser/py/imports"
	"github.com/safedep/dry/log"
	"github.com/safedep/vet/pkg/common/logger"
	"github.com/spf13/cobra"
)

var input_file string

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// findDirectDeps()
		log.Debugf("Running Scan cmd..")
	},
}

// scanCmd represents the scan command
var cmdScanFile = &cobra.Command{
	Use:   "file",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Running Scan File..")
		scanFile()
	},
}

// scanCmd represents the scan command
var cmdDirectDeps = &cobra.Command{
	Use:   "find-direct-deps",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

	rootPkgs, _ := parser.FindDirectDependencies(ctx, filename, true, includeExtensions, excludeDirs)
	for _, k := range rootPkgs.GetPackagesNames() {
		fmt.Println(k)
	}
}

func scanFile() {
	filename := input_file

	ctx := context.Background()
	cf := imports.NewPyCodeParserFactory()
	parser, err := cf.NewCodeParser()
	if err != nil {
		logger.Warnf("Error while creating parser %v", err)
		return
	}

	parsedCode, err := parser.ParseFile(ctx, filename)
	if err != nil {
		panic(err)
	}

	parsedCode.ExtractModules()
}
