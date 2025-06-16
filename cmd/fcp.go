package cmd

import (
	"cutlass/fcp"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var fcpCmd = &cobra.Command{
	Use:   "fcp",
	Short: "FCPXML generation tools",
	Long: `FCPXML generation tools for creating Final Cut Pro XML files.

This command provides various subcommands for generating and working with FCPXML files.
Use 'cutlass fcp --help' to see all available subcommands.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help when called without subcommands
		cmd.Help()
	},
}

var createEmptyCmd = &cobra.Command{
	Use:   "create-empty [filename]",
	Short: "Generate an empty FCPXML file from structs",
	Long:  `Generate a basic empty FCPXML file structure using the fcp package structs.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get output filename from flag or generate default
		output, _ := cmd.Flags().GetString("output")
		var filename string
		
		if output != "" {
			filename = output
		} else if len(args) > 0 {
			filename = args[0]
		} else {
			// Generate default filename with unix timestamp
			timestamp := time.Now().Unix()
			filename = fmt.Sprintf("cutlass_%d.fcpxml", timestamp)
		}
		
		_, err := fcp.GenerateEmpty(filename)
		if err != nil {
			fmt.Printf("Error generating FCPXML: %v\n", err)
			return
		}
		fmt.Printf("Generated empty FCPXML: %s\n", filename)
	},
}

var addVideoCmd = &cobra.Command{
	Use:   "add-video [video-file]",
	Short: "Add a video to an FCPXML file using structs",
	Long:  `Add a video asset and asset-clip to an FCPXML file using the fcp package structs.
If --input is specified, the video will be appended to an existing FCPXML file.
Otherwise, a new FCPXML file is created.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		videoFile := args[0]
		
		// Get input and output filenames from flags
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		var filename string
		
		if output != "" {
			filename = output
		} else {
			// Generate default filename with unix timestamp
			timestamp := time.Now().Unix()
			filename = fmt.Sprintf("cutlass_%d.fcpxml", timestamp)
		}
		
		var fcpxml *fcp.FCPXML
		var err error
		
		// Load existing FCPXML or create new one
		if input != "" {
			fcpxml, err = fcp.ReadFromFile(input)
			if err != nil {
				fmt.Printf("Error reading FCPXML file '%s': %v\n", input, err)
				return
			}
			fmt.Printf("Loaded existing FCPXML: %s\n", input)
		} else {
			// Generate empty FCPXML structure
			fcpxml, err = fcp.GenerateEmpty("")
			if err != nil {
				fmt.Printf("Error creating FCPXML structure: %v\n", err)
				return
			}
		}
		
		// Add video to the structure
		err = fcp.AddVideo(fcpxml, videoFile)
		if err != nil {
			fmt.Printf("Error adding video: %v\n", err)
			return
		}
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		if input != "" {
			fmt.Printf("Added video to existing FCPXML and saved to: %s\n", filename)
		} else {
			fmt.Printf("Generated FCPXML with video: %s\n", filename)
		}
	},
}

var addImageCmd = &cobra.Command{
	Use:   "add-image [image-file]",
	Short: "Add an image to an FCPXML file using structs",
	Long:  `Add an image asset and asset-clip to an FCPXML file using the fcp package structs. Supports PNG, JPG, and JPEG files.
If --input is specified, the image will be appended to an existing FCPXML file.
Otherwise, a new FCPXML file is created.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageFile := args[0]
		
		// Get duration from flag (default 9 seconds)
		durationStr, _ := cmd.Flags().GetString("duration")
		duration, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			fmt.Printf("Error parsing duration '%s': %v\n", durationStr, err)
			return
		}
		
		// Get slide animation flag
		withSlide, _ := cmd.Flags().GetBool("with-slide")
		
		// Get input and output filenames from flags
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		var filename string
		
		if output != "" {
			filename = output
		} else {
			// Generate default filename with unix timestamp
			timestamp := time.Now().Unix()
			filename = fmt.Sprintf("cutlass_%d.fcpxml", timestamp)
		}
		
		var fcpxml *fcp.FCPXML
		
		// Load existing FCPXML or create new one
		if input != "" {
			fcpxml, err = fcp.ReadFromFile(input)
			if err != nil {
				fmt.Printf("Error reading FCPXML file '%s': %v\n", input, err)
				return
			}
			fmt.Printf("Loaded existing FCPXML: %s\n", input)
		} else {
			// Generate empty FCPXML structure
			fcpxml, err = fcp.GenerateEmpty("")
			if err != nil {
				fmt.Printf("Error creating FCPXML structure: %v\n", err)
				return
			}
		}
		
		// Add image to the structure
		err = fcp.AddImageWithSlide(fcpxml, imageFile, duration, withSlide)
		if err != nil {
			fmt.Printf("Error adding image: %v\n", err)
			return
		}
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		if input != "" {
			fmt.Printf("Added image to existing FCPXML and saved to: %s (duration: %.1fs)\n", filename, duration)
		} else {
			fmt.Printf("Generated FCPXML with image: %s (duration: %.1fs)\n", filename, duration)
		}
	},
}

var addTextCmd = &cobra.Command{
	Use:   "add-text [text-file]",
	Short: "Add staggered text elements from a file to an FCPXML",
	Long:  `Add multiple text elements from a text file to an FCPXML file. Each line in the text file becomes a text element with progressive Y positioning and staggered timing.
The first text element starts at the specified offset, and each subsequent element appears 6 seconds later with a 300px Y offset.
If --input is specified, the text elements will be appended to an existing FCPXML file.
Otherwise, a new FCPXML file is created.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		textFile := args[0]
		
		// Get offset from flag (default 1 second)
		offsetStr, _ := cmd.Flags().GetString("offset")
		offset, err := strconv.ParseFloat(offsetStr, 64)
		if err != nil {
			fmt.Printf("Error parsing offset '%s': %v\n", offsetStr, err)
			return
		}
		
		// Get duration from flag (default 9 seconds)
		durationStr, _ := cmd.Flags().GetString("duration")
		duration, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			fmt.Printf("Error parsing duration '%s': %v\n", durationStr, err)
			return
		}
		
		// Get input and output filenames from flags
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		var filename string
		
		if output != "" {
			filename = output
		} else {
			// Generate default filename with unix timestamp
			timestamp := time.Now().Unix()
			filename = fmt.Sprintf("cutlass_%d.fcpxml", timestamp)
		}
		
		var fcpxml *fcp.FCPXML
		
		// Load existing FCPXML or create new one
		if input != "" {
			fcpxml, err = fcp.ReadFromFile(input)
			if err != nil {
				fmt.Printf("Error reading FCPXML file '%s': %v\n", input, err)
				return
			}
			fmt.Printf("Loaded existing FCPXML: %s\n", input)
		} else {
			// Generate empty FCPXML structure
			fcpxml, err = fcp.GenerateEmpty("")
			if err != nil {
				fmt.Printf("Error creating FCPXML structure: %v\n", err)
				return
			}
		}
		
		// Add text elements to the structure
		err = fcp.AddTextFromFile(fcpxml, textFile, offset, duration)
		if err != nil {
			fmt.Printf("Error adding text elements: %v\n", err)
			return
		}
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		if input != "" {
			fmt.Printf("Added text elements to existing FCPXML and saved to: %s (offset: %.1fs, duration: %.1fs)\n", filename, offset, duration)
		} else {
			fmt.Printf("Generated FCPXML with text elements: %s (offset: %.1fs, duration: %.1fs)\n", filename, offset, duration)
		}
	},
}

var addSlideCmd = &cobra.Command{
	Use:   "add-slide [offset]",
	Short: "Add slide animation to video at specified offset",
	Long:  `Add slide animation to the video found at the specified offset time.
The video will slide from left to right over 1 second starting from its beginning.
If the video at the offset is an AssetClip, it will be converted to a Video element to support animation.
Requires an existing FCPXML file with video content.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		offsetStr := args[0]
		
		// Parse offset
		offset, err := strconv.ParseFloat(offsetStr, 64)
		if err != nil {
			fmt.Printf("Error parsing offset '%s': %v\n", offsetStr, err)
			return
		}
		
		// Get input and output filenames from flags
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		
		if input == "" {
			fmt.Printf("Error: --input is required for add-slide command\n")
			return
		}
		
		var filename string
		if output != "" {
			filename = output
		} else {
			// Generate default filename with unix timestamp
			timestamp := time.Now().Unix()
			filename = fmt.Sprintf("cutlass_%d.fcpxml", timestamp)
		}
		
		// Load existing FCPXML
		fcpxml, err := fcp.ReadFromFile(input)
		if err != nil {
			fmt.Printf("Error reading FCPXML file '%s': %v\n", input, err)
			return
		}
		fmt.Printf("Loaded existing FCPXML: %s\n", input)
		
		// Add slide animation to video at offset
		err = fcp.AddSlideToVideoAtOffset(fcpxml, offset)
		if err != nil {
			fmt.Printf("Error adding slide animation: %v\n", err)
			return
		}
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		fmt.Printf("Added slide animation to video at offset %.1fs and saved to: %s\n", offset, filename)
	},
}

var addAudioCmd = &cobra.Command{
	Use:   "add-audio [audio-file]",
	Short: "Add an audio file as the main audio track starting at 00:00",
	Long:  `Add an audio asset and asset-clip to an FCPXML file as the main audio track starting at 00:00.
Supports WAV, MP3, M4A, and other audio formats.
If --input is specified, the audio will be added to an existing FCPXML file.
Otherwise, a new FCPXML file is created.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		audioFile := args[0]
		
		// Get input and output filenames from flags
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		var filename string
		
		if output != "" {
			filename = output
		} else {
			// Generate default filename with unix timestamp
			timestamp := time.Now().Unix()
			filename = fmt.Sprintf("cutlass_%d.fcpxml", timestamp)
		}
		
		var fcpxml *fcp.FCPXML
		var err error
		
		// Load existing FCPXML or create new one
		if input != "" {
			fcpxml, err = fcp.ReadFromFile(input)
			if err != nil {
				fmt.Printf("Error reading FCPXML file '%s': %v\n", input, err)
				return
			}
			fmt.Printf("Loaded existing FCPXML: %s\n", input)
		} else {
			// Generate empty FCPXML structure
			fcpxml, err = fcp.GenerateEmpty("")
			if err != nil {
				fmt.Printf("Error creating FCPXML structure: %v\n", err)
				return
			}
		}
		
		// Add audio to the structure
		err = fcp.AddAudio(fcpxml, audioFile)
		if err != nil {
			fmt.Printf("Error adding audio: %v\n", err)
			return
		}
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		if input != "" {
			fmt.Printf("Added audio to existing FCPXML and saved to: %s\n", filename)
		} else {
			fmt.Printf("Generated FCPXML with audio: %s\n", filename)
		}
	},
}

func init() {
	// Add output flag to create-empty subcommand
	createEmptyCmd.Flags().StringP("output", "o", "", "Output filename (defaults to cutlass_unixtime.fcpxml)")
	
	// Add flags to add-video subcommand
	addVideoCmd.Flags().StringP("input", "i", "", "Input FCPXML file to append to (optional)")
	addVideoCmd.Flags().StringP("output", "o", "", "Output filename (defaults to cutlass_unixtime.fcpxml)")
	
	// Add flags to add-image subcommand
	addImageCmd.Flags().StringP("input", "i", "", "Input FCPXML file to append to (optional)")
	addImageCmd.Flags().StringP("output", "o", "", "Output filename (defaults to cutlass_unixtime.fcpxml)")
	addImageCmd.Flags().StringP("duration", "d", "9", "Duration in seconds (default 9)")
	addImageCmd.Flags().Bool("with-slide", false, "Add keyframe animation to slide the image from left to right over 1 second")
	
	// Add flags to add-text subcommand
	addTextCmd.Flags().StringP("input", "i", "", "Input FCPXML file to append to (optional)")
	addTextCmd.Flags().StringP("output", "o", "", "Output filename (defaults to cutlass_unixtime.fcpxml)")
	addTextCmd.Flags().StringP("offset", "t", "1", "Start time offset in seconds (default 1)")
	addTextCmd.Flags().StringP("duration", "d", "9", "Duration of each text element in seconds (default 9)")
	
	// Add flags to add-slide subcommand
	addSlideCmd.Flags().StringP("input", "i", "", "Input FCPXML file to read from (required)")
	addSlideCmd.Flags().StringP("output", "o", "", "Output filename (defaults to cutlass_unixtime.fcpxml)")
	
	// Add flags to add-audio subcommand
	addAudioCmd.Flags().StringP("input", "i", "", "Input FCPXML file to append to (optional)")
	addAudioCmd.Flags().StringP("output", "o", "", "Output filename (defaults to cutlass_unixtime.fcpxml)")
	
	fcpCmd.AddCommand(createEmptyCmd)
	fcpCmd.AddCommand(addVideoCmd)
	fcpCmd.AddCommand(addImageCmd)
	fcpCmd.AddCommand(addTextCmd)
	fcpCmd.AddCommand(addSlideCmd)
	fcpCmd.AddCommand(addAudioCmd)
}