package cmd

import (
	"cutlass/fcp"
	"fmt"
	"os"
	"strconv"
	"strings"
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

var addPipVideoCmd = &cobra.Command{
	Use:   "add-pip-video [pip-video-file]",
	Short: "Add a picture-in-picture video to an existing FCPXML file",
	Long: `Add a video as picture-in-picture (PIP) to an existing FCPXML file.
The PIP video will be added as a nested asset-clip inside the first video element with:
- Proper positioning and scaling transforms
- Shape mask filter for rounded corners
- Lane "-1" for proper layering
- Optional offset timing

Requires an existing FCPXML file with at least one video element to nest the PIP inside.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pipVideoFile := args[0]
		
		// Get offset from flag (default 0 seconds)
		offsetStr, _ := cmd.Flags().GetString("offset")
		offset, err := strconv.ParseFloat(offsetStr, 64)
		if err != nil {
			fmt.Printf("Error parsing offset '%s': %v\n", offsetStr, err)
			return
		}
		
		// Get input and output filenames from flags
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		
		if input == "" {
			fmt.Printf("Error: --input is required for add-pip-video command\n")
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
		
		// Add PIP video to the structure
		err = fcp.AddPipVideo(fcpxml, pipVideoFile, offset)
		if err != nil {
			fmt.Printf("Error adding PIP video: %v\n", err)
			return
		}
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		fmt.Printf("Added PIP video to existing FCPXML and saved to: %s (offset: %.1fs)\n", filename, offset)
	},
}

var addTxtCmd = &cobra.Command{
	Use:   "add-txt [new-text]",
	Short: "Add a text message like samples/imessage001.fcpxml or append like imessage002.fcpxml",
	Long: `Add a text message to an FCPXML file recreating structures from samples/imessage001.fcpxml or imessage002.fcpxml.

For new files (no --input):
- Creates complete imessage001 structure with phone background and blue speech bubble
- Uses provided text or "Hey u there?" as default

For appending to existing files (with --input):
- Auto-detects conversation pattern and alternates bubble types
- Blue bubble (white text) for sender messages
- White bubble (black text) for reply messages  
- Continues conversation naturally without requiring original text

Examples:
  cutlass fcp add-txt                                    # Creates new with "Hey u there?"
  cutlass fcp add-txt "Hello there"                      # Creates new with custom text
  cutlass fcp add-txt "Yes, I'm here." -i existing.fcpxml  # Appends reply (auto-detects pattern)
  cutlass fcp add-txt "Got it!" -i conversation.fcpxml     # Continues alternating pattern`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var textContent string
		if len(args) > 0 {
			textContent = args[0]
		} else {
			textContent = "Hey u there?" // Default from samples/imessage001.fcpxml
		}
		
		// Get offset from flag (default 1 second)
		offsetStr, _ := cmd.Flags().GetString("offset")
		offset, err := strconv.ParseFloat(offsetStr, 64)
		if err != nil {
			fmt.Printf("Error parsing offset '%s': %v\n", offsetStr, err)
			return
		}
		
		// Get duration from flag (default 3 seconds)
		durationStr, _ := cmd.Flags().GetString("duration")
		duration, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			fmt.Printf("Error parsing duration '%s': %v\n", durationStr, err)
			return
		}
		
		// Get original-text flag for manual conversation control
		originalText, _ := cmd.Flags().GetString("original-text")
		
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
		
       // Handle appending vs creating new
       if input != "" {
           // Appending mode - read existing FCPXML
           fcpxml, err = fcp.ReadFromFile(input)
           if err != nil {
               fmt.Printf("Error reading FCPXML file '%s': %v\n", input, err)
               return
           }
           fmt.Printf("Loaded existing FCPXML: %s\n", input)

           // Append new text using appropriate method
           if originalText != "" {
               // Manual control: use AddImessageReply with specific original text
               err = fcp.AddImessageReply(fcpxml, originalText, textContent, offset, duration)
           } else {
               // Auto-detect: use AddImessageContinuation for automatic pattern detection
               err = fcp.AddImessageContinuation(fcpxml, textContent, offset, duration)
           }
           if err != nil {
               fmt.Printf("Error adding message: %v\n", err)
               return
           }
       } else {
           // Creating new mode
           fcpxml, err = fcp.GenerateEmpty("")
           if err != nil {
               fmt.Printf("Error creating FCPXML structure: %v\n", err)
               return
           }

           // Add initial text to the structure
           err = fcp.AddImessageText(fcpxml, textContent, offset, duration)
           if err != nil {
               fmt.Printf("Error adding text: %v\n", err)
               return
           }
       }
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		if input != "" {
			fmt.Printf("Added text to existing FCPXML and saved to: %s (offset: %.1fs, duration: %.1fs)\n", filename, offset, duration)
		} else {
			fmt.Printf("Generated FCPXML with text: %s (offset: %.1fs, duration: %.1fs)\n", filename, offset, duration)
		}
	},
}

var addConversationCmd = &cobra.Command{
	Use:   "add-conversation [conversation-file]",
	Short: "Create a conversation from a text file with one message per line",
	Long: `Create an iMessage-style conversation from a text file.

Each line becomes a message in the conversation.
Messages automatically alternate between blue (sender) and white (reply) bubbles.
Empty lines are ignored.

Examples:
  cutlass fcp add-conversation messages.txt -o conversation.fcpxml

File format (messages.txt):
  Hey u there?
  Yes, I'm here.
  u sure?
  i am very sure.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conversationFile := args[0]
		
		// Get flags
		output, _ := cmd.Flags().GetString("output")
		offsetStr, _ := cmd.Flags().GetString("offset")
		durationStr, _ := cmd.Flags().GetString("duration")
		
		// Parse offset and duration
		offset, err := strconv.ParseFloat(offsetStr, 64)
		if err != nil {
			fmt.Printf("Error parsing offset '%s': %v\n", offsetStr, err)
			return
		}
		duration, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			fmt.Printf("Error parsing duration '%s': %v\n", durationStr, err)
			return
		}
		
		// Read conversation file
		content, err := os.ReadFile(conversationFile)
		if err != nil {
			fmt.Printf("Error reading conversation file '%s': %v\n", conversationFile, err)
			return
		}
		
		// Parse messages (one per line, skip empty lines)
		lines := strings.Split(string(content), "\n")
		var messages []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				messages = append(messages, line)
			}
		}
		
		if len(messages) == 0 {
			fmt.Printf("No messages found in conversation file '%s'\n", conversationFile)
			return
		}
		
		// Generate output filename if not provided
		var filename string
		if output != "" {
			filename = output
		} else {
			timestamp := time.Now().Unix()
			filename = fmt.Sprintf("conversation_%d.fcpxml", timestamp)
		}
		
		// Create first message (blue bubble)
		fcpxml, err := fcp.GenerateEmpty("")
		if err != nil {
			fmt.Printf("Error creating FCPXML structure: %v\n", err)
			return
		}
		
		err = fcp.AddImessageText(fcpxml, messages[0], offset, duration)
		if err != nil {
			fmt.Printf("Error adding first message: %v\n", err)
			return
		}
		
		// Use the EXACT pattern from your working manual commands:
		// Every other message uses AddImessageReply with --original-text
		// The rest try AddImessageContinuation even though it's broken
		for i := 1; i < len(messages); i++ {
			if i%2 == 1 {
				// Odd messages (1,3,5...): white bubbles with --original-text
				fmt.Printf("DEBUG: Message %d ('%s') -> AddImessageReply\n", i+1, messages[i])
				err = fcp.AddImessageReply(fcpxml, messages[i-1], messages[i], offset, duration)
			} else {
				// Even messages (2,4,6...): attempt blue bubbles without --original-text
				// This matches: ./cutlass fcp add-txt -i trip.fcpxml "u sure?" -o trip.fcpxml
				// Even though AddImessageContinuation is broken, use it to match your sequence
				fmt.Printf("DEBUG: Message %d ('%s') -> AddImessageContinuation\n", i+1, messages[i])
				err = fcp.AddImessageContinuation(fcpxml, messages[i], offset, duration)
			}
			
			if err != nil {
				fmt.Printf("Error adding message %d ('%s'): %v\n", i+1, messages[i], err)
				return
			}
		}
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		fmt.Printf("Generated conversation FCPXML with %d messages: %s\n", len(messages), filename)
	},
}

var baffleCmd = &cobra.Command{
	Use:   "baffle [output-filename]",
	Short: "Generate a random complex 3-9 minute timeline to stress-test FCPXML generation",
	Long: `Generate a random complex timeline between 3-9 minutes containing:
- Multiple PNG and video assets from ./assets directory
- Multi-lane composition with 3-8 lanes
- Text elements with various fonts and styles
- Complex animations and transforms
- Multiple effects and filters
- Crop adjustments and spatial transformations

This command is designed to stress-test the FCPXML generation system and discover
edge cases. Each run generates a completely different timeline with random:
- Duration between 180-540 seconds (3-9 minutes)
- Asset selection and placement
- Animation timing and parameters
- Effect combinations
- Lane assignments and nesting

Use this to find and fix FCPXML generation bugs by creating complex scenarios
that might not occur in normal usage.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get output filename
		var filename string
		if len(args) > 0 {
			filename = args[0]
		} else {
			timestamp := time.Now().Unix()
			filename = fmt.Sprintf("baffle_%d.fcpxml", timestamp)
		}
		
		// Get duration range from flags
		minDurationStr, _ := cmd.Flags().GetString("min-duration")
		maxDurationStr, _ := cmd.Flags().GetString("max-duration")
		
		minDuration, err := strconv.ParseFloat(minDurationStr, 64)
		if err != nil {
			fmt.Printf("Error parsing min-duration '%s': %v\n", minDurationStr, err)
			return
		}
		
		maxDuration, err := strconv.ParseFloat(maxDurationStr, 64)
		if err != nil {
			fmt.Printf("Error parsing max-duration '%s': %v\n", maxDurationStr, err)
			return
		}
		
		if minDuration >= maxDuration {
			fmt.Printf("Error: min-duration (%.1f) must be less than max-duration (%.1f)\n", minDuration, maxDuration)
			return
		}
		
		// Get verbose flag
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		// Generate random complex timeline
		fmt.Printf("Generating random complex timeline (%.1f-%.1f minutes)...\n", minDuration/60, maxDuration/60)
		
		fcpxml, err := fcp.GenerateBaffleTimeline(minDuration, maxDuration, verbose)
		if err != nil {
			fmt.Printf("Error generating baffle timeline: %v\n", err)
			return
		}
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		fmt.Printf("Generated complex baffle timeline: %s\n", filename)
		fmt.Printf("Import this into Final Cut Pro to test for crashes and issues.\n")
	},
}

var storyBaffleCmd = &cobra.Command{
	Use:   "story-baffle [output-filename]",
	Short: "Generate an AI video creation blooper reel - story of chaos and recovery",
	Long: `Generate an engaging story about an AI trying to make a complex video and everything going wrong!

Creates a 5-minute narrative timeline containing:
- Story phases: Simple start → Getting excited → Peak chaos → Collapse → Recovery
- Multi-lane complex animations that build up and fall apart dramatically
- Dynamic text that gets bigger and crazier as the AI loses control
- Images flying in from all directions, then crashing down
- Pixabay images themed to match each story phase
- Never any black screens - always something happening!

The narrative arc:
1. "Let me make a simple video..." (calm beginning)
2. "Actually, let me add more images!" (excitement building)
3. "MOAR LAYERS! MOAR EFFECTS!" (complete chaos)
4. "Oh no! Everything is falling!" (dramatic collapse)
5. "Well, that was educational!" (quiet recovery)

This combines the storytelling of 'story' with the complex multi-lane chaos of 'baffle'
to create something that's both technically impressive AND entertaining to watch.

Examples:
  cutlass fcp story-baffle                    # Generate story_baffle_timestamp.fcpxml
  cutlass fcp story-baffle chaos.fcpxml       # Custom filename  
  cutlass fcp story-baffle --duration 300 --images 50 --verbose`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get output filename
		var filename string
		if len(args) > 0 {
			filename = args[0]
		} else {
			timestamp := time.Now().Unix()
			filename = fmt.Sprintf("story_baffle_%d.fcpxml", timestamp)
		}
		
		// Get flags
		durationStr, _ := cmd.Flags().GetString("duration")
		imagesStr, _ := cmd.Flags().GetString("images")
		complexityStr, _ := cmd.Flags().GetString("complexity")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		apiKey, _ := cmd.Flags().GetString("api-key")
		format, _ := cmd.Flags().GetString("format")
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		// Parse duration
		duration, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			fmt.Printf("Error parsing duration '%s': %v\n", durationStr, err)
			return
		}
		
		// Parse total images
		totalImages, err := strconv.Atoi(imagesStr)
		if err != nil {
			fmt.Printf("Error parsing images '%s': %v\n", imagesStr, err)
			return
		}
		
		// Parse complexity
		complexity, err := strconv.ParseFloat(complexityStr, 64)
		if err != nil {
			fmt.Printf("Error parsing complexity '%s': %v\n", complexityStr, err)
			return
		}
		
		// Validate format parameter
		if format != "horizontal" && format != "vertical" {
			fmt.Printf("Error: format must be 'horizontal' or 'vertical', got '%s'\n", format)
			return
		}
		
		// Validate complexity
		if complexity < 0.0 || complexity > 1.0 {
			fmt.Printf("Error: complexity must be between 0.0 and 1.0, got %.2f\n", complexity)
			return
		}
		
		// Create story-baffle configuration
		config := &fcp.StoryBaffleConfig{
			Duration:      duration,
			OutputDir:     outputDir,
			PixabayAPIKey: apiKey,
			MaxComplexity: complexity,
			ImageCount:    totalImages,
			Format:        format,
		}
		
		// Generate story-baffle timeline
		fmt.Printf("Generating AI video creation story-baffle (%.1f minutes)...\n", duration/60)
		
		fcpxml, err := fcp.GenerateStoryBaffle(config, verbose)
		if err != nil {
			fmt.Printf("Error generating story-baffle timeline: %v\n", err)
			return
		}
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		fmt.Printf("Generated AI video creation story-baffle: %s\n", filename)
		fmt.Printf("Images saved to: %s\n", config.OutputDir)
		fmt.Printf("Import this into Final Cut Pro for a wild ride!\n")
	},
}

var storyCmd = &cobra.Command{
	Use:   "story [output-filename]",
	Short: "Generate a 3-minute story video using random words and Pixabay images",
	Long: `Generate a narrative video using random English words and corresponding images from Pixabay.

Creates a 3-minute timeline containing:
- ~90 images downloaded from Pixabay based on random English words
- Each word generates 3 images from Pixabay search
- Images are arranged sequentially with slide animations
- Automatic duration calculation to fit 3-minute target

The story generation process:
1. Generates random English words (nature, adventure, family, etc.)
2. Searches Pixabay for 3 images per word using the public API
3. Downloads images to ./story_assets directory
4. Creates FCPXML timeline with proper image timing and animations
5. Each image gets slide animation for visual continuity

Examples:
  cutlass fcp story                          # Generate story_timestamp.fcpxml
  cutlass fcp story mystory.fcpxml           # Custom filename
  cutlass fcp story --duration 300 --images 150  # 5 minutes with 150 images
  cutlass fcp story --api-key YOUR_KEY --verbose # Use Pixabay API key`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get output filename
		var filename string
		if len(args) > 0 {
			filename = args[0]
		} else {
			timestamp := time.Now().Unix()
			filename = fmt.Sprintf("story_%d.fcpxml", timestamp)
		}
		
		// Get flags
		durationStr, _ := cmd.Flags().GetString("duration")
		imagesStr, _ := cmd.Flags().GetString("images")
		imagesPerWordStr, _ := cmd.Flags().GetString("images-per-word")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		apiKey, _ := cmd.Flags().GetString("api-key")
		showAttribution, _ := cmd.Flags().GetBool("attribution")
		attributionOutput, _ := cmd.Flags().GetString("attribution-output")
		inputFile, _ := cmd.Flags().GetString("input-file")
		format, _ := cmd.Flags().GetString("format")
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		// Parse duration
		duration, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			fmt.Printf("Error parsing duration '%s': %v\n", durationStr, err)
			return
		}
		
		// Parse total images
		totalImages, err := strconv.Atoi(imagesStr)
		if err != nil {
			fmt.Printf("Error parsing images '%s': %v\n", imagesStr, err)
			return
		}
		
		// Parse images per word
		imagesPerWord, err := strconv.Atoi(imagesPerWordStr)
		if err != nil {
			fmt.Printf("Error parsing images-per-word '%s': %v\n", imagesPerWordStr, err)
			return
		}
		
		// Validate format parameter
		if format != "horizontal" && format != "vertical" {
			fmt.Printf("Error: format must be 'horizontal' or 'vertical', got '%s'\n", format)
			return
		}
		
		// Create story configuration
		config := &fcp.StoryConfig{
			Duration:         duration,
			ImagesPerWord:    imagesPerWord,
			TotalImages:      totalImages,
			OutputDir:        outputDir,
			PixabayAPIKey:    apiKey,
			ShowAttribution:  showAttribution,
			AttributionOutput: attributionOutput,
			InputFile:        inputFile,
			Format:           format,
		}
		
		// Generate story timeline
		fmt.Printf("Generating story timeline (%.1f minutes with %d images)...\n", duration/60, totalImages)
		
		fcpxml, err := fcp.GenerateStoryTimeline(config, verbose)
		if err != nil {
			fmt.Printf("Error generating story timeline: %v\n", err)
			return
		}
		
		// Write to file
		err = fcp.WriteToFile(fcpxml, filename)
		if err != nil {
			fmt.Printf("Error writing FCPXML: %v\n", err)
			return
		}
		
		fmt.Printf("Generated story timeline: %s\n", filename)
		fmt.Printf("Images saved to: %s\n", config.OutputDir)
		fmt.Printf("Import this into Final Cut Pro to view your story.\n")
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
	
	// Add flags to add-pip-video subcommand
	addPipVideoCmd.Flags().StringP("input", "i", "", "Input FCPXML file to read from (required)")
	addPipVideoCmd.Flags().StringP("output", "o", "", "Output filename (defaults to cutlass_unixtime.fcpxml)")
	addPipVideoCmd.Flags().StringP("offset", "t", "0", "Start offset in seconds for PIP video (default 0)")
	
	// Add flags to add-txt subcommand
	addTxtCmd.Flags().StringP("input", "i", "", "Input FCPXML file to append to (optional)")
	addTxtCmd.Flags().StringP("output", "o", "", "Output filename (defaults to cutlass_unixtime.fcpxml)")
	addTxtCmd.Flags().StringP("offset", "t", "1", "Start offset in seconds for text (default 1)")
	addTxtCmd.Flags().StringP("duration", "d", "3", "Duration of text element in seconds (default 3)")
	addTxtCmd.Flags().String("original-text", "", "Original bubble text for manual control (optional - auto-detects if not provided)")
	
	// Add flags to add-conversation subcommand
	addConversationCmd.Flags().StringP("output", "o", "", "Output filename (defaults to conversation_unixtime.fcpxml)")
	addConversationCmd.Flags().StringP("offset", "t", "1", "Start offset in seconds for each message (default 1)")
	addConversationCmd.Flags().StringP("duration", "d", "3", "Duration of each message in seconds (default 3)")
	
	// Add flags to baffle subcommand
	baffleCmd.Flags().String("min-duration", "180", "Minimum timeline duration in seconds (default 180 = 3 minutes)")
	baffleCmd.Flags().String("max-duration", "540", "Maximum timeline duration in seconds (default 540 = 9 minutes)")
	baffleCmd.Flags().BoolP("verbose", "v", false, "Verbose output showing generation details")
	
	// Add flags to story-baffle subcommand
	storyBaffleCmd.Flags().String("duration", "300", "Total story duration in seconds (default 300 = 5 minutes)")
	storyBaffleCmd.Flags().String("images", "50", "Total number of images to download and use (default 50)")
	storyBaffleCmd.Flags().String("complexity", "0.95", "Maximum chaos complexity from 0.0 to 1.0 (default 0.95)")
	storyBaffleCmd.Flags().String("output-dir", "./story_baffle_assets", "Directory to save downloaded images (default ./story_baffle_assets)")
	storyBaffleCmd.Flags().String("api-key", "", "Pixabay API key for higher rate limits (optional)")
	storyBaffleCmd.Flags().String("format", "horizontal", "Video format: 'horizontal' (1280x720) or 'vertical' (1080x1920) (default 'horizontal')")
	storyBaffleCmd.Flags().BoolP("verbose", "v", false, "Verbose output showing generation details")

	// Add flags to story subcommand
	storyCmd.Flags().String("duration", "180", "Total story duration in seconds (default 180 = 3 minutes)")
	storyCmd.Flags().String("images", "90", "Total number of images to download and use (default 90)")
	storyCmd.Flags().String("images-per-word", "3", "Number of images to download per word (default 3)")
	storyCmd.Flags().String("output-dir", "./story_assets", "Directory to save downloaded images (default ./story_assets)")
	storyCmd.Flags().String("api-key", "", "Pixabay API key for higher rate limits (optional)")
	storyCmd.Flags().Bool("attribution", true, "Show attribution text for Pixabay images (default true)")
	storyCmd.Flags().String("attribution-output", "video", "Where to output attribution: 'video' (text elements), 'stdout' (console), 'both', or 'none' (default 'video')")
	storyCmd.Flags().String("input-file", "", "Text file with sentences (one per line) to use instead of random words")
	storyCmd.Flags().String("format", "horizontal", "Video format: 'horizontal' (1280x720) or 'vertical' (1080x1920) (default 'horizontal')")
	storyCmd.Flags().BoolP("verbose", "v", false, "Verbose output showing generation details")
	
	fcpCmd.AddCommand(createEmptyCmd)
	fcpCmd.AddCommand(addVideoCmd)
	fcpCmd.AddCommand(addImageCmd)
	fcpCmd.AddCommand(addTextCmd)
	fcpCmd.AddCommand(addSlideCmd)
	fcpCmd.AddCommand(addAudioCmd)
	fcpCmd.AddCommand(addPipVideoCmd)
	fcpCmd.AddCommand(addTxtCmd)
	fcpCmd.AddCommand(addConversationCmd)
	fcpCmd.AddCommand(baffleCmd)
	fcpCmd.AddCommand(storyBaffleCmd)
	fcpCmd.AddCommand(storyCmd)
}
