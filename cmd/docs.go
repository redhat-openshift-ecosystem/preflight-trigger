/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"log"
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation for the project",
	Long:  ``,
	Run:   docsRun,
}

func init() {
	createCmd.AddCommand(docsCmd)
	docsCmd.Flags().StringVarP(&CommandFlags.DocsType, "docs-type", "", "markdown", "Type of documentation to generate. Supported types: man, markdown, rest, yaml")
}

func docsRun(cmd *cobra.Command, args []string) {
	doctypes := map[string]interface{}{
		"man":      doc.GenManTree,
		"markdown": doc.GenMarkdownTree,
		"rest":     doc.GenReSTTree,
		"yaml":     doc.GenYamlTree,
	}

	if CommandFlags.DocsType == "man" {
		header := &doc.GenManHeader{Title: "PREFLIGHT TRIGGER", Section: "5"}
		docsGenerator := doctypes[CommandFlags.DocsType].(func(cmd *cobra.Command, header *doc.GenManHeader, dir string) error)
		if err := docsGenerator(rootCmd, header, "docs"); err != nil {
			log.Fatalf("Unable to generate docs: %v", err)
		}
	} else {
		docsGenerator := doctypes[CommandFlags.DocsType].(func(cmd *cobra.Command, dir string) error)
		if err := docsGenerator(rootCmd, "docs"); err != nil {
			log.Fatalf("Unable to generate docs: %v", err)
		}
	}
}
