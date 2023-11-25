/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/safedep/codex/pkg/parser"
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
		fmt.Println("scan called")
		scan()
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.PersistentFlags().StringVar(&input_file, "input", "", "Provide  Github Acc Name")
	scanCmd.MarkPersistentFlagRequired("input")

}

func scan() {
	filename := input_file

	ctx := context.Background()
	cf := parser.CodeSnippetFactory{}
	parsedCode, err := cf.ParseCode(ctx, filename)
	if err != nil {
		logger.Warnf("Error while creating parser %v", err)
		return
	}

	// parsedCode.Analyze()
	parsedCode.Query(`
	(import_statement
		name: (dotted_name 
				(identifier) @import_module
		)
	)
	
	(import_from_statement
		 module_name: (dotted_name) @from_module
		 name: (dotted_name 
				(identifier) @import_submodule
		)
	)
	
	(import_from_statement
		 module_name: (relative_import) @import_prefix
		 name: (dotted_name 
				(identifier) @import_submodule
		)
	)	
	
	`)
}
