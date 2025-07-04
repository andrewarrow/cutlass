package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cutlass",
	Short: "A Swiss Army knife for generating FCPXML files",
	Long: `Cutlass is a powerful CLI tool for generating FCPXML files from various sources.
It provides a comprehensive set of commands organized into logical categories to help
you create Final Cut Pro XML files for video editing workflows.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(utilsCmd)
	rootCmd.AddCommand(fcpCmd)
}
