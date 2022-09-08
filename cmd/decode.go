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

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Decode a value or local file; value or file location is required",
	Long: `Decode accepts either a value or a local file to decode and print to stdout.
If you set the output-path the decoded data will be written to the specified file.`,
	PreRun: decodePreRun,
	Run:    decodeRun,
}

func init() {
	rootCmd.AddCommand(decodeCmd)
	decodeCmd.Flags().StringVarP(&CommandFlags.ValueToDecode, "value", "", "", "Value to decode")
	decodeCmd.Flags().StringVarP(&CommandFlags.FileToDecode, "file", "", "", "Path of file to decode")
	decodeCmd.Flags().BoolVarP(&CommandFlags.UseStdin, "stdin", "", false, "Read value from stdin")
	decodeCmd.MarkFlagsMutuallyExclusive("value", "file")

}

func decodePreRun(cmd *cobra.Command, args []string) {
	if CommandFlags.ValueToDecode == "" && CommandFlags.FileToDecode == "" && !CommandFlags.UseStdin {
		if err := cmd.Help(); err != nil {
			return
		}
		os.Exit(1)
	}
}

func decodeRun(cmd *cobra.Command, args []string) {
	var decodedData string

	if CommandFlags.UseStdin {
		CommandFlags.ValueToDecode = getStdin()
	}

	if CommandFlags.ValueToDecode != "" {
		decodedData = decode(CommandFlags.ValueToDecode)
	} else {
		fileToDecode, err := os.ReadFile(CommandFlags.FileToDecode)
		if err != nil {
			log.Fatalf("Unable to read file for decoding: %v", err)
		}
		decodedData = decode(string(fileToDecode))
	}

	if CommandFlags.OutputPath == "" {
		_, err := os.Stdout.WriteString(decodedData)
		if err != nil {
			log.Fatalf("Unable to write decoded data to stdout: %v", err)
		}
	} else {
		err := os.WriteFile(CommandFlags.OutputPath, []byte(decodedData), 0644)
		if err != nil {
			log.Fatalf("Unable to write decoded data to file: %v", err)
		}
	}
}

func decode(valueToDecode string) string {
	data, err := base64.StdEncoding.DecodeString(valueToDecode)
	if err != nil {
		log.Fatalf("Unable to decode data: %v", err)
	}

	return string(data)
}
