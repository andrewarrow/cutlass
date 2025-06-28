package wikipedia

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type WikipediaAPIResponse struct {
	Query struct {
		Random []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"random"`
	} `json:"query"`
}

// HandleWikipediaRandomCommand processes random Wikipedia articles
func HandleWikipediaRandomCommand(args []string, max int) {
	fmt.Printf("Fetching %d random Wikipedia articles...\n", max)

	// Create data directory if it doesn't exist
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating data directory: %v\n", err)
		return
	}

	successCount := 0
	for i := 0; i < max; i++ {
		fmt.Printf("\n=== Processing article %d/%d ===\n", i+1, max)
		
		// Get random article title
		title, err := getRandomWikipediaTitle()
		if err != nil {
			fmt.Printf("Error getting random article: %v\n", err)
			continue
		}

		fmt.Printf("Processing: %s\n", title)

		// Generate FCPXML from the article
		outputFile := filepath.Join(dataDir, fmt.Sprintf("wiki_%d_%s.fcpxml", i+1, sanitizeFilename(title)))
		
		if err := GenerateFromWikipedia(title, outputFile); err != nil {
			fmt.Printf("Error generating FCPXML for '%s': %v\n", title, err)
			continue
		}

		successCount++
		fmt.Printf("Successfully generated: %s\n", outputFile)

		// Add small delay between requests to be respectful
		if i < max-1 {
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Printf("\nCompleted processing %d out of %d articles successfully!\n", successCount, max)
}

// getRandomWikipediaTitle fetches a random Wikipedia article title using the API
func getRandomWikipediaTitle() (string, error) {
	// Use Wikipedia API to get random article
	apiURL := "https://en.wikipedia.org/api/rest_v1/page/random/summary"
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch random article: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var result struct {
		Title string `json:"title"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if result.Title == "" {
		return "", fmt.Errorf("no title found in response")
	}

	return result.Title, nil
}

// sanitizeFilename creates a safe filename from a title
func sanitizeFilename(title string) string {
	// Remove or replace characters that aren't safe for filenames
	safe := strings.ReplaceAll(title, " ", "_")
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	safe = strings.ReplaceAll(safe, ":", "_")
	safe = strings.ReplaceAll(safe, "*", "_")
	safe = strings.ReplaceAll(safe, "?", "_")
	safe = strings.ReplaceAll(safe, "\"", "_")
	safe = strings.ReplaceAll(safe, "<", "_")
	safe = strings.ReplaceAll(safe, ">", "_")
	safe = strings.ReplaceAll(safe, "|", "_")
	
	// Convert to lowercase
	safe = strings.ToLower(safe)

	// Limit length to avoid filesystem issues
	if len(safe) > 100 {
		safe = safe[:100]
	}

	return safe
}