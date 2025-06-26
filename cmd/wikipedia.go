package cmd

import (
	"cutlass/wikipedia"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var wikipediaCmd = &cobra.Command{
	Use:   "wikipedia [article-title]",
	Short: "Generate FCPXML from Wikipedia articles and tables",
	Long: `Generate FCPXML files from Wikipedia articles and tables.
This command allows you to extract tables from Wikipedia articles and convert them
to FCPXML format for use in Final Cut Pro.

If you provide an article title directly (not random, table, or parse), it will
generate FCPXML from the article's tables, just like the 'table' subcommand.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no args provided, show help
		if len(args) == 0 {
			cmd.Help()
			return
		}

		articleTitle := args[0]
		
		// Check if it's one of the subcommands
		if articleTitle == "table" || articleTitle == "parse" || articleTitle == "random" {
			// Let subcommands handle this
			cmd.Help()
			return
		}

		// Default behavior: generate FCPXML from article (like old code)
		outputFile, _ := cmd.Flags().GetString("output")
		
		// If no output file specified, use article title as filename
		if outputFile == "" {
			outputFile = articleTitle + ".fcpxml"
		} else if !strings.HasSuffix(strings.ToLower(outputFile), ".fcpxml") {
			outputFile += ".fcpxml"
		}

		fmt.Printf("Using Wikipedia mode to create FCPXML from article tables...\n")
		if err := wikipedia.GenerateFromWikipedia(articleTitle, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating from Wikipedia: %v\n", err)
			os.Exit(1)
		}
	},
}

var wikipediaTableCmd = &cobra.Command{
	Use:   "table <article-title>",
	Short: "Extract and generate FCPXML from Wikipedia article tables",
	Long: `Extract tables from a Wikipedia article and generate FCPXML.
This command fetches the Wikipedia article, parses its tables, and creates
an FCPXML file suitable for Final Cut Pro.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		articleTitle := args[0]
		outputFile, _ := cmd.Flags().GetString("output")
		
		// If no output file specified, use article title as filename
		if outputFile == "" {
			outputFile = articleTitle + ".fcpxml"
		} else if !strings.HasSuffix(strings.ToLower(outputFile), ".fcpxml") {
			outputFile += ".fcpxml"
		}

		fmt.Printf("Using Wikipedia mode to create FCPXML from article tables...\n")
		if err := wikipedia.GenerateFromWikipedia(articleTitle, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating from Wikipedia: %v\n", err)
			os.Exit(1)
		}
	},
}

var wikipediaParseCmd = &cobra.Command{
	Use:   "parse <article-title>",
	Short: "Parse and display Wikipedia article tables",
	Long: `Parse Wikipedia article tables and display them in ASCII format.
This command allows you to preview the table structure before generating FCPXML.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		articleTitle := args[0]
		tableNumber, _ := cmd.Flags().GetInt("table-num")
		
		if err := wikipedia.ParseWikipediaTables(articleTitle, tableNumber); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing Wikipedia tables: %v\n", err)
			os.Exit(1)
		}
	},
}

var wikipediaRandomCmd = &cobra.Command{
	Use:   "random",
	Short: "Generate content from random Wikipedia articles",
	Long: `Generate content from random Wikipedia articles.
This command fetches random Wikipedia articles, extracts content, and creates
media files with FCPXML output.`,
	Run: func(cmd *cobra.Command, args []string) {
		maxStr, _ := cmd.Flags().GetString("max")
		max := 10 // default
		if maxStr != "" {
			if parsed, err := strconv.Atoi(maxStr); err == nil {
				max = parsed
			}
		}
		
		wikipedia.HandleWikipediaRandomCommand(args, max)
	},
}

func init() {
	// Add Wikipedia command to root
	rootCmd.AddCommand(wikipediaCmd)
	
	// Add subcommands to Wikipedia command
	wikipediaCmd.AddCommand(wikipediaTableCmd)
	wikipediaCmd.AddCommand(wikipediaParseCmd)
	wikipediaCmd.AddCommand(wikipediaRandomCmd)
	
	// Add flags for main wikipedia command (for direct article generation)
	wikipediaCmd.Flags().StringP("output", "o", "", "Output file")
	
	// Add flags for table command
	wikipediaTableCmd.Flags().StringP("output", "o", "", "Output file")
	
	// Add flags for parse command
	wikipediaParseCmd.Flags().IntP("table-num", "t", 0, "Table number to display (0 for all, 1-N for specific table)")
	
	// Add flags for random command
	wikipediaRandomCmd.Flags().StringP("max", "m", "10", "Maximum number of articles to process")
}