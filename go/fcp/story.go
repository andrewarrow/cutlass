// Package fcp provides FCPXML generation for story creation.
//
// Story generation creates narrative videos using random English words
// and corresponding images from Pixabay. Each word becomes a visual element
// in a 3-minute story timeline.
package fcp

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	rand_math "math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Common English words for story generation
var englishWords = []string{
	"adventure", "animal", "beautiful", "bird", "car", "cat", "city", "cloud", "color",
	"dance", "dog", "dream", "earth", "family", "fire", "flower", "forest", "friend",
	"garden", "happy", "heart", "home", "house", "journey", "light", "love", "mountain",
	"nature", "ocean", "peace", "people", "rainbow", "river", "smile", "sun", "sunset",
	"tree", "water", "weather", "wind", "winter", "wonder", "world", "beach", "bridge",
	"castle", "children", "door", "freedom", "gold", "green", "hero", "island", "magic",
	"moon", "music", "night", "path", "quiet", "road", "sky", "snow", "star", "storm",
	"summer", "time", "travel", "village", "wave", "whisper", "art", "book", "butterfly",
	"childhood", "curiosity", "discovery", "emotion", "energy", "excitement", "exploration",
	"flower", "growth", "happiness", "imagination", "inspiration", "joy", "knowledge",
	"laughter", "memory", "mystery", "opportunity", "passion", "photography", "playground",
	"possibility", "reflection", "serenity", "strength", "surprise", "transformation",
	"victory", "wisdom", "young", "adventure", "beauty", "courage", "determination",
}

// PixabayResponse represents the structure of Pixabay API response
type PixabayResponse struct {
	Hits []PixabayHit `json:"hits"`
}

// PixabayHit represents a single image result from Pixabay
type PixabayHit struct {
	ID           int    `json:"id"`
	WebformatURL string `json:"webformatURL"`
	Tags         string `json:"tags"`
	User         string `json:"user"`         // Photographer/creator display name
	UserID       int    `json:"user_id"`     // Photographer user ID
	Type         string `json:"type"`        // photo, illustration, vector
	Category     string `json:"category"`    // nature, backgrounds, etc.
	Views        int    `json:"views"`       // Number of views
	Downloads    int    `json:"downloads"`   // Number of downloads
	Likes        int    `json:"likes"`       // Number of likes
}

// ImageAttribution holds attribution information for downloaded images
type ImageAttribution struct {
	FilePath string // Local file path
	Source   string // "pixabay" or "lorem"
	Author   string // Author/photographer name (empty for Lorem Picsum)
	UserID   int    // Pixabay user ID (0 for Lorem Picsum)
	PixabayID int   // Original Pixabay image ID (0 for Lorem Picsum)
}

// StoryConfig holds configuration for story generation
type StoryConfig struct {
	Duration         float64 // Total duration in seconds (default: 180 = 3 minutes)
	ImagesPerWord    int     // Number of images to download per word (default: 3)
	TotalImages      int     // Target total number of images (default: 90)
	OutputDir        string  // Directory to store downloaded images
	PixabayAPIKey    string  // Pixabay API key (optional, uses public API if empty)
	ShowAttribution  bool    // Whether to show attribution text for Pixabay images (default: true)
	AttributionOutput string  // Where to output attribution: "video", "stdout", "both", or "none" (default: "video")
	InputFile        string  // Path to text file with sentences (one per line) to use instead of random words
	Format           string  // Video format: "horizontal" (1280x720) or "vertical" (1080x1920) (default: "horizontal")
}

// DefaultStoryConfig returns a default configuration for story generation
func DefaultStoryConfig() *StoryConfig {
	return &StoryConfig{
		Duration:         180.0, // 3 minutes
		ImagesPerWord:    3,
		TotalImages:      90,
		OutputDir:        "./story_assets",
		ShowAttribution:  true,   // Enable attribution by default
		AttributionOutput: "video", // Default to video text elements
		Format:           "horizontal", // Default to horizontal format
	}
}

// generateRandomFilename creates a random UUID-like filename
func generateRandomFilename() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateRandomWords generates a list of random English words
func GenerateRandomWords(count int) []string {
	rand_math.Seed(time.Now().UnixNano())
	
	words := make([]string, count)
	for i := 0; i < count; i++ {
		words[i] = englishWords[rand_math.Intn(len(englishWords))]
	}
	
	return words
}

// ReadSentencesFromFile reads sentences from a text file, one per line
func ReadSentencesFromFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filepath, err)
	}
	defer file.Close()

	var sentences []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" { // Skip empty lines
			sentences = append(sentences, line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filepath, err)
	}
	
	if len(sentences) == 0 {
		return nil, fmt.Errorf("no sentences found in file %s", filepath)
	}
	
	return sentences, nil
}

// HighContrastColor represents a face color and its high-contrast outline color
type HighContrastColor struct {
	FaceColor    string // RGBA format "r g b a"
	OutlineColor string // RGBA format "r g b a"
	Name         string // Human-readable name for debugging
}

// GetRandomHighContrastColors returns a list of high-contrast color combinations
func GetRandomHighContrastColors() []HighContrastColor {
	return []HighContrastColor{
		{FaceColor: "1 1 1 1", OutlineColor: "0 0 0 1", Name: "White on Black"},       // White text, black outline
		{FaceColor: "0 0 0 1", OutlineColor: "1 1 1 1", Name: "Black on White"},       // Black text, white outline
		{FaceColor: "1 0 0 1", OutlineColor: "1 1 1 1", Name: "Red on White"},         // Red text, white outline
		{FaceColor: "0 1 0 1", OutlineColor: "0 0 0 1", Name: "Green on Black"},       // Green text, black outline
		{FaceColor: "0 0 1 1", OutlineColor: "1 1 1 1", Name: "Blue on White"},        // Blue text, white outline
		{FaceColor: "1 1 0 1", OutlineColor: "0 0 0 1", Name: "Yellow on Black"},      // Yellow text, black outline
		{FaceColor: "1 0 1 1", OutlineColor: "1 1 1 1", Name: "Magenta on White"},     // Magenta text, white outline
		{FaceColor: "0 1 1 1", OutlineColor: "0 0 0 1", Name: "Cyan on Black"},        // Cyan text, black outline
		{FaceColor: "1 0.5 0 1", OutlineColor: "0 0 0 1", Name: "Orange on Black"},    // Orange text, black outline
		{FaceColor: "0.5 0 1 1", OutlineColor: "1 1 1 1", Name: "Purple on White"},    // Purple text, white outline
	}
}

// GetRandomFonts returns a list of fonts to choose from, loading from reference/fonts.txt
func GetRandomFonts() []string {
	fonts, err := loadFontsFromFile()
	if err != nil {
		// Fallback to hardcoded fonts if file reading fails
		return []string{
			"Sinzano",
			"Helvetica Neue",
			"Arial",
			"Times New Roman",
			"Courier New",
			"Verdana",
			"Georgia",
			"Trebuchet MS",
			"Comic Sans MS",
			"Impact",
		}
	}
	return fonts
}

// loadFontsFromFile loads the complete font list from reference/fonts.txt
func loadFontsFromFile() ([]string, error) {
	// Try to read from reference/fonts.txt
	content, err := os.ReadFile("reference/fonts.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read fonts file: %v", err)
	}
	
	lines := strings.Split(string(content), "\n")
	var fonts []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			fonts = append(fonts, line)
		}
	}
	
	if len(fonts) == 0 {
		return nil, fmt.Errorf("no fonts found in file")
	}
	
	return fonts, nil
}

// DownloadImagesFromPixabay downloads images for a given word from Pixabay or fallback sources
func DownloadImagesFromPixabay(word string, count int, outputDir string, apiKey string) ([]ImageAttribution, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}
	
	// Try Pixabay first if API key is provided
	if apiKey != "" {
		if files, err := downloadFromPixabay(word, count, outputDir, apiKey); err == nil {
			return files, nil
		}
	}
	
	// Fallback to Lorem Picsum with themed seeds based on word
	return downloadFromLoremPicsum(word, count, outputDir)
}

// downloadFromPixabay attempts to download from Pixabay API
func downloadFromPixabay(word string, count int, outputDir string, apiKey string) ([]ImageAttribution, error) {
	// Build Pixabay API URL
	baseURL := "https://pixabay.com/api/"
	params := url.Values{}
	params.Add("q", word)
	params.Add("key", apiKey)
	params.Add("image_type", "photo")
	params.Add("orientation", "horizontal")
	params.Add("category", "all")
	params.Add("min_width", "640")
	params.Add("min_height", "480")
	// Pixabay API requires per_page to be between 3 and 200
	perPage := count
	if perPage < 3 {
		perPage = 3
	}
	if perPage > 200 {
		perPage = 200
	}
	params.Add("per_page", fmt.Sprintf("%d", perPage))
	params.Add("safesearch", "true")
	
	requestURL := baseURL + "?" + params.Encode()
	
	// Create HTTP client with 3 second timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	
	// Make HTTP request to Pixabay API
	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch images from Pixabay: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		// Read response body for debugging
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Pixabay API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	// Parse JSON response
	var pixabayResp PixabayResponse
	if err := json.NewDecoder(resp.Body).Decode(&pixabayResp); err != nil {
		return nil, fmt.Errorf("failed to parse Pixabay response: %v", err)
	}
	
	
	if len(pixabayResp.Hits) == 0 {
		return nil, fmt.Errorf("no images found for word: %s", word)
	}
	
	// Download images
	var downloadedFiles []ImageAttribution
	for i, hit := range pixabayResp.Hits {
		if i >= count {
			break
		}
		
		// Download image with random UUID filename to prevent UID conflicts
		uuidStr := generateRandomFilename()
		filename := fmt.Sprintf("%s.jpg", uuidStr)
		filepath := filepath.Join(outputDir, filename)
		
		if err := downloadImage(hit.WebformatURL, filepath); err != nil {
			fmt.Printf("Warning: Failed to download image %s: %v\n", hit.WebformatURL, err)
			continue
		}
		
		// Create attribution info
		attribution := ImageAttribution{
			FilePath:  filepath,
			Source:    "pixabay",
			Author:    hit.User,
			UserID:    hit.UserID, 
			PixabayID: hit.ID,
		}
		
		downloadedFiles = append(downloadedFiles, attribution)
	}
	
	if len(downloadedFiles) == 0 {
		return nil, fmt.Errorf("failed to download any images for word: %s", word)
	}
	
	return downloadedFiles, nil
}

// downloadFromLoremPicsum downloads placeholder images from Lorem Picsum
func downloadFromLoremPicsum(word string, count int, outputDir string) ([]ImageAttribution, error) {
	var downloadedFiles []ImageAttribution
	
	// Create a simple hash from the word to get consistent images
	hash := 0
	for _, char := range word {
		hash = hash*31 + int(char)
	}
	if hash < 0 {
		hash = -hash
	}
	
	for i := 0; i < count; i++ {
		// Generate a seed based on word hash and index
		seed := (hash + i*137) % 1000 // Keep seed within reasonable range
		
		// Lorem Picsum URL with seed for consistent images
		imageURL := fmt.Sprintf("https://picsum.photos/seed/%s%d/800/600", word, seed)
		
		// Download image with random UUID filename to prevent UID conflicts
		uuidStr := generateRandomFilename()
		filename := fmt.Sprintf("%s.jpg", uuidStr)
		filepath := filepath.Join(outputDir, filename)
		
		if err := downloadImage(imageURL, filepath); err != nil {
			fmt.Printf("Warning: Failed to download image %s: %v\n", imageURL, err)
			continue
		}
		
		// Create attribution info for Lorem Picsum (no author)
		attribution := ImageAttribution{
			FilePath:  filepath,
			Source:    "lorem",
			Author:    "", // No author for Lorem Picsum
			UserID:    0,  // No user ID for Lorem Picsum
			PixabayID: 0,  // No Pixabay ID for Lorem Picsum
		}
		
		downloadedFiles = append(downloadedFiles, attribution)
	}
	
	if len(downloadedFiles) == 0 {
		return nil, fmt.Errorf("failed to download any images for word: %s", word)
	}
	
	return downloadedFiles, nil
}

// downloadImage downloads an image from a URL to a local file
func downloadImage(imageURL, filepath string) error {
	// Create HTTP client with 3 second timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	
	resp, err := client.Get(imageURL)
	if err != nil {
		return fmt.Errorf("failed to fetch image: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("image request returned status %d", resp.StatusCode)
	}
	
	// Create output file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()
	
	// Copy image data to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write image data: %v", err)
	}
	
	return nil
}

// GenerateStoryTimeline creates a 3-minute story timeline using random words and Pixabay images
func GenerateStoryTimeline(config *StoryConfig, verbose bool) (*FCPXML, error) {
	if config == nil {
		config = DefaultStoryConfig()
	}
	
	// Create base FCPXML structure with specified format
	fcpxml, err := GenerateEmptyWithFormat("", config.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to create base FCPXML: %v", err)
	}
	
	// Set up resource management
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()
	
	if verbose {
		fmt.Printf("Starting story timeline generation...\n")
		fmt.Printf("Target duration: %.1f seconds (%.1f minutes)\n", config.Duration, config.Duration/60)
		fmt.Printf("Target images: %d\n", config.TotalImages)
	}
	
	// Get words/sentences based on input source
	var words []string
	if config.InputFile != "" {
		// Read sentences from input file
		sentences, err := ReadSentencesFromFile(config.InputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read input file: %v", err)
		}
		words = sentences
		if verbose {
			fmt.Printf("Read %d sentences from file: %s\n", len(words), config.InputFile)
			fmt.Printf("First few sentences: %s\n", strings.Join(words[:min(3, len(words))], " | "))
			if len(words) > 3 {
				fmt.Printf("... and %d more\n", len(words)-3)
			}
		}
	} else {
		// Calculate how many words we need for random generation
		wordsNeeded := config.TotalImages / config.ImagesPerWord
		if config.TotalImages%config.ImagesPerWord != 0 {
			wordsNeeded++
		}
		
		// Generate random words
		words = GenerateRandomWords(wordsNeeded)
		if verbose {
			fmt.Printf("Generated %d words: %s\n", len(words), strings.Join(words[:min(5, len(words))], ", "))
			if len(words) > 5 {
				fmt.Printf("... and %d more\n", len(words)-5)
			}
		}
	}
	
	// Download images for each word/sentence
	var allImageAttributions []ImageAttribution
	for i, searchTerm := range words {
		searchType := "word"
		if config.InputFile != "" {
			searchType = "sentence"
		}
		if verbose {
			fmt.Printf("Downloading images for %s %d/%d: %s\n", searchType, i+1, len(words), searchTerm)
		}
		
		imageAttributions, err := DownloadImagesFromPixabay(searchTerm, config.ImagesPerWord, config.OutputDir, config.PixabayAPIKey)
		if err != nil {
			fmt.Printf("Warning: Failed to download images for %s '%s': %v\n", searchType, searchTerm, err)
			continue
		}
		
		allImageAttributions = append(allImageAttributions, imageAttributions...)
		
		// Stop if we have enough images
		if len(allImageAttributions) >= config.TotalImages {
			allImageAttributions = allImageAttributions[:config.TotalImages]
			break
		}
	}
	
	if len(allImageAttributions) == 0 {
		return nil, fmt.Errorf("no images were downloaded successfully")
	}
	
	if verbose {
		fmt.Printf("Downloaded %d images total\n", len(allImageAttributions))
	}
	
	// Generate timeline with images and text overlays
	imageDuration := config.Duration / float64(len(allImageAttributions))
	wordIndex := 0
	
	for i, imageAttr := range allImageAttributions {
		// Add image with proper duration and format, passing image index for alternating Ken Burns direction
		err := AddImageWithSlideAndFormatIndex(fcpxml, imageAttr.FilePath, imageDuration, true, config.Format, i)
		if err != nil {
			fmt.Printf("Warning: Failed to add image %s: %v\n", imageAttr.FilePath, err)
			continue
		}
		
		// Add text overlay for corresponding word/sentence (one per images-per-word images)
		if i%config.ImagesPerWord == 0 && wordIndex < len(words) {
			textOffset := float64(i) * imageDuration
			textContent := words[wordIndex]
			
			// Use smaller font size for sentences to fit better
			fontSize := 290
			if config.InputFile != "" {
				fontSize = 150 // Smaller font for sentences
			}
			
			// Add text with appropriate font size and format
			err = AddStoryTextWithFormat(fcpxml, textContent, textOffset, imageDuration, fontSize, config.Format)
			if err != nil {
				if verbose {
					textType := "word"
					if config.InputFile != "" {
						textType = "sentence"
					}
					fmt.Printf("Warning: Failed to add %s '%s' at offset %.1fs: %v\n", textType, textContent, textOffset, err)
				}
			} else if verbose {
				textType := "word"
				if config.InputFile != "" {
					textType = "sentence"
				}
				fmt.Printf("Added %s '%s' at offset %.1fs\n", textType, textContent, textOffset)
			}
			
			wordIndex++
		}
		
		// Handle attribution output based on configuration
		if imageAttr.Source == "pixabay" && imageAttr.Author != "" {
			imageOffset := float64(i) * imageDuration
			attributionText := fmt.Sprintf("https://pixabay.com/users/%s-%d/", strings.ToLower(imageAttr.Author), imageAttr.UserID)
			
			// Output to stdout if requested
			if config.AttributionOutput == "stdout" || config.AttributionOutput == "both" {
				fmt.Printf("Attribution: %s (for image: %s)\n", attributionText, imageAttr.FilePath)
			}
			
			// Add to video if requested (default behavior for backward compatibility)
			shouldAddToVideo := config.ShowAttribution && (config.AttributionOutput == "video" || config.AttributionOutput == "both")
			if shouldAddToVideo {
				err = AddAttributionText(fcpxml, attributionText, imageOffset, imageDuration)
				if err != nil {
					if verbose {
						fmt.Printf("Warning: Failed to add attribution '%s' at offset %.1fs: %v\n", attributionText, imageOffset, err)
					}
				} else if verbose {
					fmt.Printf("Added attribution '%s' at offset %.1fs\n", attributionText, imageOffset)
				}
			}
		}
		
		if verbose && (i+1)%10 == 0 {
			fmt.Printf("Added %d/%d images to timeline\n", i+1, len(allImageAttributions))
		}
	}
	
	// Update sequence duration
	updateSequenceDuration(fcpxml, config.Duration)
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}
	
	if verbose {
		fmt.Printf("Story timeline generation completed successfully\n")
		fmt.Printf("Final timeline duration: %.1f seconds with %d images\n", config.Duration, len(allImageAttributions))
	}
	
	return fcpxml, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AddStoryText adds a single text element to the story timeline with specified font size
func AddStoryText(fcpxml *FCPXML, text string, offsetSeconds float64, durationSeconds float64, fontSize int) error {
	return AddStoryTextWithFormat(fcpxml, text, offsetSeconds, durationSeconds, fontSize, "horizontal")
}

// AddStoryTextWithFormat adds a single text element to the story timeline with specified font size and format
func AddStoryTextWithFormat(fcpxml *FCPXML, text string, offsetSeconds float64, durationSeconds float64, fontSize int, format string) error {
	// Use the existing resource registry
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	// Find or create text effect
	textEffectID := ""
	for _, effect := range fcpxml.Resources.Effects {
		if strings.Contains(effect.UID, "Text.moti") {
			textEffectID = effect.ID
			break
		}
	}

	if textEffectID == "" {
		ids := tx.ReserveIDs(1)
		textEffectID = ids[0]
		
		_, err := tx.CreateEffect(textEffectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
		if err != nil {
			return fmt.Errorf("failed to create text effect: %v", err)
		}
	}

	// Generate unique text style ID
	textStyleID := GenerateTextStyleID(text, fmt.Sprintf("story_text_offset_%.1f", offsetSeconds))
	
	// Select random colors and font
	colors := GetRandomHighContrastColors()
	fonts := GetRandomFonts()
	
	rand_math.Seed(time.Now().UnixNano() + int64(offsetSeconds*1000)) // Ensure different seed for each text
	selectedColor := colors[rand_math.Intn(len(colors))]
	selectedFont := fonts[rand_math.Intn(len(fonts))]
	
	// Output selected font to stdout
	fmt.Printf("Text: \"%s\" -> Font: %s\n", text, selectedFont)
	
	// Convert durations to FCP format
	offsetDuration := ConvertSecondsToFCPDuration(offsetSeconds)
	titleDuration := ConvertSecondsToFCPDuration(durationSeconds)

	// Find the target video/asset-clip first to get the correct offset
	var targetVideo *Video
	var targetAssetClip *AssetClip
	var titleOffset string = offsetDuration // default fallback

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		offsetFrames := parseFCPDuration(offsetDuration)

		// Find the video/asset-clip that covers this time offset
		for i := range sequence.Spine.AssetClips {
			clip := &sequence.Spine.AssetClips[i]
			clipOffsetFrames := parseFCPDuration(clip.Offset)
			clipDurationFrames := parseFCPDuration(clip.Duration)
			clipEndFrames := clipOffsetFrames + clipDurationFrames

			if offsetFrames >= clipOffsetFrames && offsetFrames < clipEndFrames {
				targetAssetClip = clip
				titleOffset = clip.Start // Use video start time as offset!
				break
			}
		}

		if targetAssetClip == nil {
			for i := range sequence.Spine.Videos {
				video := &sequence.Spine.Videos[i]
				videoOffsetFrames := parseFCPDuration(video.Offset)
				videoDurationFrames := parseFCPDuration(video.Duration)
				videoEndFrames := videoOffsetFrames + videoDurationFrames

				if offsetFrames >= videoOffsetFrames && offsetFrames < videoEndFrames {
					targetVideo = video
					titleOffset = video.Start // Use video start time as offset!
					break
				}
			}
		}
	}

	// Create title with large font (290 size like baffle)
	title := Title{
		Ref:      textEffectID,
		Lane:     "2", // Use lane 2 for text overlay
		Offset:   titleOffset, // Use video start time as offset (key fix!)
		Name:     text + " - Story Text",
		Start:    "86486400/24000s",
		Duration: titleDuration,
		Params: []Param{
			{
				Name:  "Position",
				Key:   "9999/10003/13260/3296672360/1/100/101", 
				Value: "0 0", // Center position
			},
			{
				Name:  "Layout Method",
				Key:   "9999/10003/13260/3296672360/2/314",
				Value: "1 (Paragraph)",
			},
			{
				Name:  "Left Margin",
				Key:   "9999/10003/13260/3296672360/2/323",
				Value: getLeftMargin(format),
			},
			{
				Name:  "Right Margin", 
				Key:   "9999/10003/13260/3296672360/2/324",
				Value: getRightMargin(format),
			},
			{
				Name:  "Top Margin",
				Key:   "9999/10003/13260/3296672360/2/325",
				Value: getTopMargin(format),
			},
			{
				Name:  "Bottom Margin",
				Key:   "9999/10003/13260/3296672360/2/326",
				Value: getBottomMargin(format),
			},
			{
				Name:  "Alignment",
				Key:   "9999/10003/13260/3296672360/2/354/3296667315/401",
				Value: "1 (Center)", // Center alignment
			},
			{
				Name:  "Line Spacing",
				Key:   "9999/10003/13260/3296672360/2/354/3296667315/404",
				Value: "-19",
			},
			{
				Name:  "Auto-Shrink",
				Key:   "9999/10003/13260/3296672360/2/370",
				Value: "3 (To All Margins)",
			},
			{
				Name:  "Alignment",
				Key:   "9999/10003/13260/3296672360/2/373",
				Value: "1 (Center) 1 (Middle)",
			},
			// Add stroke parameters for outline effect
			{
				Name:  "Size",
				Key:   "9999/10003/13260/3296672360/5/3296672362/3",
				Value: "288", // Font size
			},
			{
				Name:  "Color",
				Key:   "9999/10003/13260/3296672360/5/3296672362/30/32",
				Value: selectedColor.OutlineColor[:len(selectedColor.OutlineColor)-2], // Remove alpha channel for stroke color
			},
			{
				Name:  "Wrap Mode",
				Key:   "9999/10003/13260/3296672360/5/3296672362/30/34/5",
				Value: "1 (Repeat)",
			},
			{
				Name:  "Width",
				Key:   "9999/10003/13260/3296672360/5/3296672362/30/36",
				Value: "15", // Stroke width 15.0
			},
			{
				Name:  "Font",
				Key:   "9999/10003/13260/3296672360/5/3296672362/83",
				Value: "206 0", // Font selection
			},
		},
		Text: &TitleText{
			TextStyles: []TextStyleRef{
				{
					Ref:  textStyleID,
					Text: text,
				},
			},
		},
		TextStyleDefs: []TextStyleDef{
			{
				ID: textStyleID,
				TextStyle: TextStyle{
					Font:         selectedFont,
					FontSize:     "288", // Always use 288 as specified
					FontFace:     "Regular",
					FontColor:    selectedColor.FaceColor,
					StrokeColor:  selectedColor.OutlineColor,
					StrokeWidth:  "-15", // Negative value for outline
					Alignment:    "center",
					LineSpacing:  "-19",
				},
			},
		},
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit text transaction: %v", err)
	}

	// Add text to the appropriate element (we already found the targets above)
	if targetAssetClip != nil {
		targetAssetClip.Titles = append(targetAssetClip.Titles, title)
	} else if targetVideo != nil {
		targetVideo.NestedTitles = append(targetVideo.NestedTitles, title)
	} else if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		// Fallback: add to spine directly
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		sequence.Spine.Titles = append(sequence.Spine.Titles, title)
	}

	return nil
}

// AddAttributionText adds small attribution text in the upper right corner for Pixabay images
func AddAttributionText(fcpxml *FCPXML, attributionText string, offsetSeconds float64, durationSeconds float64) error {
	// Use the existing resource registry
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	// Find or create text effect
	textEffectID := ""
	for _, effect := range fcpxml.Resources.Effects {
		if strings.Contains(effect.UID, "Text.moti") {
			textEffectID = effect.ID
			break
		}
	}

	if textEffectID == "" {
		ids := tx.ReserveIDs(1)
		textEffectID = ids[0]
		
		_, err := tx.CreateEffect(textEffectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
		if err != nil {
			return fmt.Errorf("failed to create text effect: %v", err)
		}
	}

	// Generate unique text style ID for attribution
	textStyleID := GenerateTextStyleID(attributionText, fmt.Sprintf("attribution_offset_%.1f", offsetSeconds))
	
	// Convert durations to FCP format
	offsetDuration := ConvertSecondsToFCPDuration(offsetSeconds)
	titleDuration := ConvertSecondsToFCPDuration(durationSeconds)

	// Find the target video/asset-clip first to get the correct offset
	var targetVideo *Video
	var targetAssetClip *AssetClip
	var titleOffset string = offsetDuration // default fallback

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		offsetFrames := parseFCPDuration(offsetDuration)

		// Find the video/asset-clip that covers this time offset
		for i := range sequence.Spine.AssetClips {
			clip := &sequence.Spine.AssetClips[i]
			clipOffsetFrames := parseFCPDuration(clip.Offset)
			clipDurationFrames := parseFCPDuration(clip.Duration)
			clipEndFrames := clipOffsetFrames + clipDurationFrames

			if offsetFrames >= clipOffsetFrames && offsetFrames < clipEndFrames {
				targetAssetClip = clip
				titleOffset = clip.Start // Use video start time as offset!
				break
			}
		}

		if targetAssetClip == nil {
			for i := range sequence.Spine.Videos {
				video := &sequence.Spine.Videos[i]
				videoOffsetFrames := parseFCPDuration(video.Offset)
				videoDurationFrames := parseFCPDuration(video.Duration)
				videoEndFrames := videoOffsetFrames + videoDurationFrames

				if offsetFrames >= videoOffsetFrames && offsetFrames < videoEndFrames {
					targetVideo = video
					titleOffset = video.Start // Use video start time as offset!
					break
				}
			}
		}
	}

	// Create attribution title with small font size and positioned in upper right
	title := Title{
		Ref:      textEffectID,
		Lane:     "3", // Use lane 3 for attribution overlay (above main text)
		Offset:   titleOffset, // Use video start time as offset
		Name:     attributionText + " - Attribution",
		Start:    "86486400/24000s",
		Duration: titleDuration,
		Params: []Param{
			{
				Name:  "Position",
				Key:   "9999/10003/13260/3296672360/1/100/101", 
				Value: "1780 1934", // Upper right position (from Info.fcpxml)
			},
			{
				Name:  "Layout Method",
				Key:   "9999/10003/13260/3296672360/2/314",
				Value: "1 (Paragraph)",
			},
			{
				Name:  "Left Margin",
				Key:   "9999/10003/13260/3296672360/2/323",
				Value: "-1500", // Wide left margin to give text horizontal space
			},
			{
				Name:  "Right Margin", 
				Key:   "9999/10003/13260/3296672360/2/324",
				Value: "50", // Small right margin from edge
			},
			{
				Name:  "Top Margin",
				Key:   "9999/10003/13260/3296672360/2/325",
				Value: "-900", // Top margin for upper positioning
			},
			{
				Name:  "Bottom Margin",
				Key:   "9999/10003/13260/3296672360/2/326",
				Value: "900", // Bottom margin to limit height
			},
			{
				Name:  "Alignment",
				Key:   "9999/10003/13260/3296672360/2/354/3296667315/401",
				Value: "2 (Right)", // Right alignment
			},
			{
				Name:  "Line Spacing",
				Key:   "9999/10003/13260/3296672360/2/354/3296667315/404",
				Value: "0", // Normal line spacing
			},
			{
				Name:  "Alignment",
				Key:   "9999/10003/13260/3296672360/2/373",
				Value: "2 (Right) 0 (Top)", // Right and top alignment
			},
		},
		Text: &TitleText{
			TextStyles: []TextStyleRef{
				{
					Ref:  textStyleID,
					Text: attributionText,
				},
			},
		},
		TextStyleDefs: []TextStyleDef{
			{
				ID: textStyleID,
				TextStyle: TextStyle{
					Font:        "Helvetica Neue",
					FontSize:    "123", // Font size from Info.fcpxml
					FontFace:    "Regular",
					FontColor:   "1 1 1 0.8", // White text with slight transparency
					Alignment:   "right",
					LineSpacing: "0",
					Bold:        "0", // Not bold for attribution
				},
			},
		},
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit attribution text transaction: %v", err)
	}

	// Add attribution text to the appropriate element
	if targetAssetClip != nil {
		targetAssetClip.Titles = append(targetAssetClip.Titles, title)
	} else if targetVideo != nil {
		targetVideo.NestedTitles = append(targetVideo.NestedTitles, title)
	} else if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		// Fallback: add to spine directly
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		sequence.Spine.Titles = append(sequence.Spine.Titles, title)
	}

	return nil
}

// Helper functions for format-specific text positioning
func getLeftMargin(format string) string {
	switch format {
	case "vertical":
		return "-970"  // Narrower margins for vertical (1080 width)
	default:
		return "-1730" // Original for horizontal (1280 width)
	}
}

func getRightMargin(format string) string {
	switch format {
	case "vertical":
		return "970"   // Narrower margins for vertical (1080 width)
	default:
		return "1730"  // Original for horizontal (1280 width)
	}
}

func getTopMargin(format string) string {
	switch format {
	case "vertical":
		return "1540"  // Higher top margin for vertical (1920 height)
	default:
		return "960"   // Original for horizontal (720 height)
	}
}

func getBottomMargin(format string) string {
	switch format {
	case "vertical":
		return "-1540" // Lower bottom margin for vertical (1920 height)
	default:
		return "-960"  // Original for horizontal (720 height)
	}
}