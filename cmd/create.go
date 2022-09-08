/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create contains subcommands for creating jobs and documentation.",
	Long: `Create should not be used alone as it does nothing. It is meant to be used
as a subcommand of the root command for you to call job or docs; see help for those 
commands for how to use them.`,
}

func init() {
	rootCmd.AddCommand(createCmd)
}
