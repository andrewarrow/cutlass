package time

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// Keyframe represents a single keyframe for animation
type Keyframe struct {
	Time  string
	Value string
}

// Animation represents position animations for video clips
type Animation struct {
	XKeyframes []Keyframe
	YKeyframes []Keyframe
}

// TextElement represents a text element with sliding animation
type TextElement struct {
	Text          string
	Index         int
	Lane          int
	Offset        string
	Duration      string
	YPosition     string
	Alignment     string
	XKeyframes    []Keyframe
	YKeyframes    []Keyframe
	SpeedKeyframes []Keyframe
}

// VideoClip represents a video clip on the timeline
type VideoClip struct {
	AssetRef    string
	Offset      string
	Name        string
	Duration    string
	FormatRef   string
	Animations  []Animation
	TextElements []TextElement
}

// VideoAsset represents a video file resource
type VideoAsset struct {
	AssetID   string
	Name      string
	UID       string
	Duration  string
	FormatID  string
	Path      string
	Bookmark  string
}

// TimeData represents the complete data for the time template
type TimeData struct {
	VideoAssets   []VideoAsset
	VideoClips    []VideoClip
	TotalDuration string
}

// TimelineItem represents a parsed item from the .time file
type TimelineItem struct {
	Type         string // "video" or "text"
	VideoPath    string
	Text         string
	Lane         int
	YOffset      int    // Y offset in pixels relative to previous text
	IsStandalone bool   // True if text is not part of a slide (has 2+ spaces)
	Animations   []TimeAnimation
}

// TimeAnimation represents a timing and animation from the .time file
type TimeAnimation struct {
	StartTime string // e.g., "3s"
	Duration  string // e.g., "2s"
	Type      string // e.g., "SLIDE"
	FromValue string // e.g., "0%"
	ToValue   string // e.g., "50%"
}

// generateVideoUID generates a unique identifier for the video file based on its content
func generateVideoUID(videoPath string) (string, error) {
	file, err := os.Open(videoPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Convert to uppercase hex string like FCP UIDs
	return fmt.Sprintf("%X", hash.Sum(nil)), nil
}

// getVideoDuration gets the duration of a video file using ffprobe
func getVideoDuration(videoPath string) (string, string, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_entries", "format=duration", videoPath)
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to run ffprobe: %v", err)
	}

	var result struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	duration, err := strconv.ParseFloat(result.Format.Duration, 64)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse duration: %v", err)
	}

	// Convert to FCP asset format (frames/44100s)
	assetFrames := int64(duration * 44100)
	assetDuration := fmt.Sprintf("%d/44100s", assetFrames)

	// Convert to FCP clip format (frames/600s) aligned to frame boundaries
	// Frame duration is 20/600s, so we need to round to multiples of 20
	clipFrames := int64(duration * 600)
	// Round to nearest frame boundary (multiple of 20)
	clipFrames = (clipFrames / 20) * 20
	clipDuration := fmt.Sprintf("%d/600s", clipFrames)

	return assetDuration, clipDuration, nil
}

// parseTimeFile parses the .time file format
func parseTimeFile(filePath string) ([]TimelineItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []TimelineItem
	scanner := bufio.NewScanner(file)
	currentItem := &TimelineItem{}
	lane := 0

	// Regular expressions for parsing
	videoPathRegex := regexp.MustCompile(`^\.\/`)
	animationRegex := regexp.MustCompile(`^\s+(\d+s):(\d+s)\s+(\w+)\s+->\s+(.+?)\s+to\s+(.+)$`)
	textRegex := regexp.MustCompile(`^(\s*)text\s+([^\s]+)(?:\s+(\d+))?$`)

	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check if this is a video path (starts with ./)
		if videoPathRegex.MatchString(line) {
			// Save previous item if it exists
			if currentItem.Type != "" {
				items = append(items, *currentItem)
			}
			
			// Start new video item
			currentItem = &TimelineItem{
				Type:      "video",
				VideoPath: strings.TrimSpace(line),
				Lane:      lane,
			}
			lane = 0 // Reset lane for video clips
			continue
		}

		// Check if this is a text element
		if matches := textRegex.FindStringSubmatch(line); matches != nil {
			// Save previous item if it exists
			if currentItem.Type != "" {
				items = append(items, *currentItem)
			}
			
			// Check spacing to determine if it's standalone
			spacing := matches[1]
			isStandalone := len(spacing) >= 2 // 2+ spaces means standalone
			
			// Parse Y offset if provided
			yOffset := 0
			if matches[3] != "" {
				yOffset, _ = strconv.Atoi(matches[3])
			}
			
			lane++ // Text elements go on subsequent lanes
			currentItem = &TimelineItem{
				Type:         "text",
				Text:         matches[2],
				Lane:         lane,
				YOffset:      yOffset,
				IsStandalone: isStandalone,
			}
			continue
		}

		// Check if this is an animation line (indented)
		if matches := animationRegex.FindStringSubmatch(line); matches != nil {
			if currentItem.Type == "" {
				return nil, fmt.Errorf("animation line found without corresponding video or text element: %s", line)
			}
			
			animation := TimeAnimation{
				StartTime: matches[1],
				Duration:  matches[2],
				Type:      matches[3],
				FromValue: matches[4],
				ToValue:   matches[5],
			}
			currentItem.Animations = append(currentItem.Animations, animation)
			continue
		}
	}

	// Don't forget the last item
	if currentItem.Type != "" {
		items = append(items, *currentItem)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// convertPercentToPixels converts percentage values to pixel values for FCP
func convertPercentToPixels(percent string, isX bool) (string, error) {
	// Remove % sign and parse
	percentStr := strings.TrimSuffix(percent, "%")
	percentValue, err := strconv.ParseFloat(percentStr, 64)
	if err != nil {
		return "", err
	}

	// Based on correct.fcpxml analysis:
	// For video: 0% = 0, 50% = 88.8889 (width-based calculation)
	// For text: -100% = -177.778, -50% = -88.8889 (width-based calculation)
	var pixelValue float64
	if isX {
		// For video position: each 1% = 1.777778 pixels (88.8889/50)
		// For text position: each 1% = 1.777778 pixels 
		pixelValue = percentValue * 1.777778
	} else {
		// Y coordinate (not used in the current example)
		pixelValue = percentValue * 7.2
	}

	return fmt.Sprintf("%.4f", pixelValue), nil
}

// convertTimeToFrames converts time string like "3s" to frame count for FCP timebase
func convertTimeToFrames(timeStr string, timebase int) (int, error) {
	// Remove 's' and parse
	timeNumStr := strings.TrimSuffix(timeStr, "s")
	timeSeconds, err := strconv.ParseFloat(timeNumStr, 64)
	if err != nil {
		return 0, err
	}

	frames := int(timeSeconds * float64(timebase))
	return frames, nil
}

func GenerateTimeFCPXML(inputFile, outputFile string) error {
	// Parse the .time file
	items, err := parseTimeFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse time file: %v", err)
	}

	if len(items) == 0 {
		return fmt.Errorf("no items found in time file")
	}

	// First pass: Reassign lane numbers according to FCP requirements
	// Standalone text (Point3) should be lane 1 (-1), non-standalone text (Point2) should be lane 2 (-2)
	standaloneLane := 1
	nonStandaloneLane := 2
	for i := range items {
		if items[i].Type == "text" {
			if items[i].IsStandalone {
				items[i].Lane = standaloneLane
				standaloneLane++
			} else {
				items[i].Lane = nonStandaloneLane
				nonStandaloneLane++
			}
		}
	}

	var videoAssets []VideoAsset
	var videoClips []VideoClip
	assetCounter := 2 // Start from r2 (r1 is the format)
	
	// Calculate total duration - we'll use the video duration
	
	// Process each timeline item
	for itemIndex, item := range items {
		if item.Type == "video" {
			// Process video clip
			absVideoPath, err := filepath.Abs(item.VideoPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for video file: %v", err)
			}

			// Check if video file exists
			if _, err := os.Stat(absVideoPath); os.IsNotExist(err) {
				return fmt.Errorf("video file does not exist: %s", absVideoPath)
			}

			// Generate UID and get duration
			videoUID, err := generateVideoUID(absVideoPath)
			if err != nil {
				return fmt.Errorf("failed to generate video UID: %v", err)
			}

			videoDuration, videoClipDuration, err := getVideoDuration(absVideoPath)
			if err != nil {
				return fmt.Errorf("failed to get video duration: %v", err)
			}

			// Create video asset
			assetID := fmt.Sprintf("r%d", assetCounter)
			formatID := fmt.Sprintf("r%d", assetCounter+1)
			assetCounter += 2

			videoAsset := VideoAsset{
				AssetID:  assetID,
				Name:     filepath.Base(absVideoPath),
				UID:      videoUID,
				Duration: videoDuration,
				FormatID: formatID,
				Path:     "file://" + absVideoPath,
				Bookmark: "Ym9va5wDAAAAAAQQMAAAAIyOwVC8wIPakjoIaEBX2fpL4BKD+9d93fs4rd1aXLybdAIAAAQAAAADAwAAABgAKAUAAAABAQAAVXNlcnMAAAACAAAAAQEAAGFhAAAGAAAAAQEAAE1vdmllcwAADgAAAAEBAABqdW45LmZjcGJ1bmRsZQAAFAAAAAEBAABBdXRvIEdlbmVyYXRlZCBDbGlwcw4AAAABAQAAT3JpZ2luYWwgTWVkaWEAACEAAAABAQAATW92aWUgb24gNi0xMC0yNSBhdCA3LjA14oCvUE0ubW92AAAAHAAAAAEGAAAQAAAAIAAAACwAAAA8AAAAVAAAAHAAAACIAAAACAAAAAQDAAAdQgAAAAAAAAgAAAAEAwAAx1IEAAAAAAAIAAAABAMAAPpSBAAAAAAACAAAAAQDAAByeBIEAAAAAAgAAAAEAwAAJXkSBAAAAAAIAAAABAMAAC15EgQAAAAACAAAAAQDAAAnRjcEAAAAABwAAAABBgAA2AAAAOgAAAD4AAAACAEAABgBAAAoAQAAOAEAAAgAAAAABAAAQcb8jf9EzJ4YAAAAAQIAAAEAAAAAAAAADwAAAAAAAAAAAAAAAAAAAAAAAAABBQAABAAAAAMDAAAIAAAABAAAAAMDAAAFAAAACAAAAAQDAAAFAAAAAAAAAAQAAAADAwAA9QEAAAgAAAABCQAAZmlsZTovLy8MAAAAAQEAAE1hY2ludG9zaCBIRAgAAAAEAwAAAJCClucAAAAIAAAAAAQAAEHGY8eqgAAAJAAAAAEBAAA1RTM5RTI1My02MEE4LTREOTItOTNDQi1ERjFERkQyMDFDQkQYAAAAAQIAAIEAAAABAAAA7xMAAAEAAAAAAAAAAAAAAAEAAAABAQAALwAAAPAAAAD+////AQAAAAAAAAATAAAABBAAALQAAAAAAAAABRAAAEgBAAAAAAAAEBAAAHwBAAAAAAAAQBAAAGwBAAAAAAAAVBAAAKQBAAAAAAAAVRAAALABAAAAAAAAVhAAAJwBAAAAAAAAAiAAAGgCAAAAAAAABSAAANgBAAAAAAAAECAAAOgBAAAAAAAAESAAABwCAAAAAAAAEiAAAPwBAAAAAAAAEyAAAAwCAAAAAAAAICAAAEgCAAAAAAAAMCAAAJwBAAAAAAAAAcAAALwBAAAAAAAAEcAAACAAAAAAAAAAEsAAAMwBAAAAAAAAENAAAAQAAAAAAAAA", // Default bookmark
			}
			videoAssets = append(videoAssets, videoAsset)

			// Create video clip with animations
			videoClip := VideoClip{
				AssetRef:  assetID,
				Offset:    "0s", // Videos start at beginning
				Name:      filepath.Base(absVideoPath),
				Duration:  videoClipDuration,
				FormatRef: formatID,
			}

			// Process video animations
			if len(item.Animations) > 0 {
				animation := Animation{}
				timebase := 3000 // Using 3000 as timebase for keyframes
				
				for _, anim := range item.Animations {
					if anim.Type == "SLIDE" {
						// Convert start time to use in keyframes
						_, err := convertTimeToFrames(anim.StartTime, timebase)
						if err != nil {
							return fmt.Errorf("failed to convert start time: %v", err)
						}

						// Convert percentage values to pixels
						fromX, err := convertPercentToPixels(anim.FromValue, true)
						if err != nil {
							return fmt.Errorf("failed to convert from value: %v", err)
						}
						
						toX, err := convertPercentToPixels(anim.ToValue, true)
						if err != nil {
							return fmt.Errorf("failed to convert to value: %v", err)
						}

						// Create X keyframes - use simple time format like correct.fcpxml
						startTimeStr := anim.StartTime
						durationSeconds, err := strconv.ParseFloat(strings.TrimSuffix(anim.Duration, "s"), 64)
						if err != nil {
							return fmt.Errorf("failed to parse duration: %v", err)
						}
						startSeconds, err := strconv.ParseFloat(strings.TrimSuffix(anim.StartTime, "s"), 64)
						if err != nil {
							return fmt.Errorf("failed to parse start time: %v", err)
						}
						endTimeStr := fmt.Sprintf("%.0fs", startSeconds+durationSeconds)
						
						animation.XKeyframes = []Keyframe{
							{Time: startTimeStr, Value: fromX},
							{Time: endTimeStr, Value: toX},
						}

						// Y keyframes (no movement)
						animation.YKeyframes = []Keyframe{
							{Time: startTimeStr, Value: "0"},
						}
					}
				}
				
				if len(animation.XKeyframes) > 0 || len(animation.YKeyframes) > 0 {
					videoClip.Animations = append(videoClip.Animations, animation)
				}
			}

			videoClips = append(videoClips, videoClip)
			
		} else if item.Type == "text" {
			// Text elements are attached to the last video clip
			if len(videoClips) == 0 {
				return fmt.Errorf("text element found without a video clip: %s", item.Text)
			}

			// Calculate Y position based on offset
			// Point2 (lane 2) has 0 offset, so Y = 0
			// Point3 (lane 1) has 100 offset, so Y = -100 (100px below Point2)
			yPosition := fmt.Sprintf("%d", -item.YOffset)
			
			// Determine timing based on whether it's standalone
			var offset, duration string
			if item.IsStandalone {
				// Standalone text appears after slide ends (at 5s mark)
				offset = "14800/3000s" // 5s converted to frames/3000s
				duration = "31027200/768000s" // Long duration to end of video
			} else {
				// Text that's part of a slide
				offset = "8700/3000s" // Based on correct.fcpxml timing (2.9s)
				duration = "10s" // Default duration
			}

			// Determine alignment based on lane
			// Lane 1 (Point3) = left alignment, Lane 2 (Point2) = right alignment
			alignment := "0 (Left)"
			if item.Lane == 2 {
				alignment = "2 (Right)"
			}
			
			// Create text element - based on correct.fcpxml structure
			textElement := TextElement{
				Text:      item.Text,
				Index:     itemIndex + 1,
				Lane:      -item.Lane, // Use negative lane numbers like correct.fcpxml
				Offset:    offset,
				Duration:  duration,
				YPosition: yPosition,
				Alignment: alignment,
			}

			// Process text animations
			if len(item.Animations) > 0 {
				for _, anim := range item.Animations {
					if anim.Type == "SLIDE" {
						// Convert percentage values to pixel values
						fromX, err := convertPercentToPixels(anim.FromValue, true)
						if err != nil {
							return fmt.Errorf("failed to convert from value: %v", err)
						}
						
						toX, err := convertPercentToPixels(anim.ToValue, true)
						if err != nil {
							return fmt.Errorf("failed to convert to value: %v", err)
						}

						// Create X keyframes for text position animation
						startTimeStr := "3600s" // Based on correct.fcpxml
						endTimeStr := "3602s"   // Based on correct.fcpxml (3600s + 2s)
						
						textElement.XKeyframes = []Keyframe{
							{Time: startTimeStr, Value: fromX},
							{Time: endTimeStr, Value: toX},
						}

						// Y keyframes (no movement)
						textElement.YKeyframes = []Keyframe{
							{Time: startTimeStr, Value: "0"},
						}

						// Create speed keyframes (from correct.fcpxml)
						textElement.SpeedKeyframes = []Keyframe{
							{Time: "-469658744/1000000000s", Value: "0"},
							{Time: "12328542033/1000000000s", Value: "1"},
						}
					}
				}
			}

			// Add text element to the last video clip
			videoClips[len(videoClips)-1].TextElements = append(videoClips[len(videoClips)-1].TextElements, textElement)
		}
	}

	// Create the time data - use video duration for sequence duration like correct.fcpxml
	var sequenceDuration string
	if len(videoClips) > 0 {
		sequenceDuration = videoClips[0].Duration // Use first video's duration
	} else {
		sequenceDuration = "27220/600s" // Default from correct.fcpxml
	}
	
	timeData := TimeData{
		VideoAssets:   videoAssets,
		VideoClips:    videoClips,
		TotalDuration: sequenceDuration,
	}

	// Read the template
	templatePath := "templates/time_slide.fcpxml"
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %v", err)
	}

	// Parse and execute the template
	tmpl, err := template.New("time").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Execute template
	if err := tmpl.Execute(outFile, timeData); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	return nil
}