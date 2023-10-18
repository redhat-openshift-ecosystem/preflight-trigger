/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/base64"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// encodeCmd represents the encode command
var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "Encode a value or local file; value or file location is required",
	Long: `Encode accepts either a value or a local file to encode and print to stdout.
If you set the output-path the encoded data will be written to the specified file.`,
	PreRun: encodePreRun,
	Run:    encodeRun,
}

func init() {
	rootCmd.AddCommand(encodeCmd)
	encodeCmd.Flags().StringVarP(&CommandFlags.ValueToEncode, "value", "", "", "Value to encode")
	encodeCmd.Flags().StringVarP(&CommandFlags.FileToEncode, "file", "", "", "Path to file to encode")
	encodeCmd.Flags().BoolVarP(&CommandFlags.UseStdin, "stdin", "", false, "Read value from stdin")
	encodeCmd.MarkFlagsMutuallyExclusive("value", "file")
}

func encodePreRun(cmd *cobra.Command, args []string) {
	if CommandFlags.ValueToEncode == "" && CommandFlags.FileToEncode == "" && !CommandFlags.UseStdin {
		if err := cmd.Help(); err != nil {
			return
		}
		os.Exit(1)
	}
}

func encodeRun(cmd *cobra.Command, args []string) {
	var encodedData string

	if CommandFlags.UseStdin {
		CommandFlags.ValueToEncode = getStdin()
	}

	if CommandFlags.ValueToEncode != "" {
		encodedData = encode(CommandFlags.ValueToEncode)
	} else {
		fileToEncode, err := os.ReadFile(CommandFlags.FileToEncode)
		if err != nil {
			log.Fatalf("Unable to read file for encoding: %v", err)
		}
		encodedData = encode(string(fileToEncode))
	}

	if CommandFlags.OutputPath == "" {
		_, err := os.Stdout.WriteString(encodedData)
		if err != nil {
			log.Fatalf("Unable to write encoded data to stdout: %v", err)
		}
	} else {
		log.Printf("Writing encoded data to file %s", CommandFlags.OutputPath)
		if err := os.WriteFile(CommandFlags.OutputPath, []byte(encodedData), 0o644); err != nil {
			log.Fatalf("Unable to write encoded data to file: %v", err)
		}
	}
}

func encode(valueToEncode string) string {
	return base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(valueToEncode))
}
