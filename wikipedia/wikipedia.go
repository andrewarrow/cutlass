package wikipedia

import (
	"cutlass/fcp"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type TableCell struct {
	Content    string
	Style      map[string]string
	Class      string
	ColSpan    int
	RowSpan    int
	Attributes map[string]string
}

type TableRow struct {
	Cells []TableCell
}

type Table struct {
	Headers []string
	Rows    []TableRow
}

type SimpleTable struct {
	Headers []string
	Rows    [][]string
}

type WikipediaData struct {
	Title  string
	Tables []Table
}

// GenerateFromWikipedia creates FCPXML from Wikipedia article tables using the new fcp system
func GenerateFromWikipedia(articleTitle, outputFile string) error {
	// Fetch Wikipedia source
	fmt.Printf("Fetching Wikipedia source for: %s\n", articleTitle)
	source, err := fetchWikipediaSource(articleTitle)
	if err != nil {
		return fmt.Errorf("failed to fetch Wikipedia source: %v", err)
	}

	// Parse the source to extract tables
	fmt.Printf("Parsing Wikipedia source for tables...\n")
	tables, err := parseWikitableFromSource(source)
	if err != nil {
		return fmt.Errorf("failed to parse Wikipedia source: %v", err)
	}

	if len(tables) == 0 {
		return fmt.Errorf("no tables found in Wikipedia article")
	}

	// Select best table
	bestTable := selectBestTable(tables)
	if bestTable == nil {
		return fmt.Errorf("no suitable table found")
	}

	fmt.Printf("Table headers: %v\n", bestTable.Headers)
	fmt.Printf("Table has %d rows\n", len(bestTable.Rows))

	// Generate FCPXML using new fcp system
	fmt.Printf("Generating FCPXML: %s\n", outputFile)
	err = generateTableFCPXML(bestTable, outputFile)
	if err != nil {
		return fmt.Errorf("failed to generate FCPXML: %v", err)
	}

	fmt.Printf("Successfully generated Wikipedia table FCPXML: %s\n", outputFile)
	return nil
}

// generateTableFCPXML creates FCPXML from table data using the new fcp system
func generateTableFCPXML(table *SimpleTable, outputFile string) error {
	// Create base FCPXML using new system
	fcpxml, err := fcp.GenerateEmpty("")
	if err != nil {
		return fmt.Errorf("failed to create base FCPXML: %v", err)
	}

	// Use proper resource management
	registry := fcp.NewResourceRegistry(fcpxml)
	tx := fcp.NewTransaction(registry)
	defer tx.Rollback()

	// Create text effect first (required for title elements)
	textEffectID := ""
	for _, effect := range fcpxml.Resources.Effects {
		if strings.Contains(effect.UID, "Text.moti") {
			textEffectID = effect.ID
			break
		}
	}

	if textEffectID == "" {
		// Reserve ID for text effect
		ids := tx.ReserveIDs(1)
		textEffectID = ids[0]

		// Create text effect using transaction
		_, err = tx.CreateEffect(textEffectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
		if err != nil {
			return fmt.Errorf("failed to create text effect: %v", err)
		}
	}

	// Create title clips for each table row
	duration := 3.0 // 3 seconds per row
	startTime := 0.0
	totalDuration := 0.0

	for i, row := range table.Rows {
		if len(row) == 0 {
			continue
		}

		// Create text content from row data
		var textParts []string
		for j, cell := range row {
			if j < len(table.Headers) && cell != "" {
				textParts = append(textParts, fmt.Sprintf("%s: %s", table.Headers[j], cell))
			}
		}
		
		if len(textParts) == 0 {
			continue
		}

		textContent := strings.Join(textParts, " | ")
		
		// Add text clip using new system
		if err := addTextClip(fcpxml, tx, textContent, startTime, duration, textEffectID); err != nil {
			return fmt.Errorf("failed to add text clip %d: %v", i+1, err)
		}

		startTime += duration
		totalDuration = startTime // Update total duration
	}

	// Update sequence duration to match total content duration
	if totalDuration > 0 && len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		sequence.Duration = fcp.ConvertSecondsToFCPDuration(totalDuration)
	}

	// Commit transaction and write
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return fcp.WriteToFile(fcpxml, outputFile)
}

// addTextClip adds a text clip to the FCPXML using the new fcp system
func addTextClip(fcpxml *fcp.FCPXML, tx *fcp.ResourceTransaction, text string, startTime, duration float64, textEffectID string) error {
	// Reserve IDs for the text style
	ids := tx.ReserveIDs(1)
	styleID := ids[0]

	// Convert times to FCP duration format
	fcpDuration := fcp.ConvertSecondsToFCPDuration(duration)
	fcpOffset := fcp.ConvertSecondsToFCPDuration(startTime)

	// Create title element with text that references the effect
	title := fcp.Title{
		Ref:      textEffectID,
		Name:     "Text",
		Offset:   fcpOffset,
		Duration: fcpDuration,
		Text: &fcp.TitleText{
			TextStyles: []fcp.TextStyleRef{
				{
					Ref:  styleID,
					Text: text,
				},
			},
		},
		TextStyleDefs: []fcp.TextStyleDef{
			{
				ID: styleID,
				TextStyle: fcp.TextStyle{
					Font:      "Helvetica",
					FontSize:  "48",
					FontColor: "1 1 1 1",
				},
			},
		},
	}

	// Add to spine
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		sequence.Spine.Titles = append(sequence.Spine.Titles, title)
	}

	return nil
}

// fetchWikipediaSource fetches the source of a Wikipedia article
func fetchWikipediaSource(articleTitle string) (string, error) {
	encodedTitle := url.QueryEscape(articleTitle)
	sourceURL := fmt.Sprintf("https://en.wikipedia.org/w/index.php?title=%s&action=edit", encodedTitle)
	
	fmt.Printf("Fetching Wikipedia source from: %s\n", sourceURL)
	
	resp, err := http.Get(sourceURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Wikipedia source: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %s", resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	
	// Extract the content from the textarea
	content := string(body)
	
	// Try different patterns for extracting the content
	patterns := []string{
		`<textarea[^>]*id="wpTextbox1"[^>]*>(.*?)</textarea>`,
		`<textarea[^>]*name="wpTextbox1"[^>]*>(.*?)</textarea>`,
		`<textarea[^>]*>(.*?)</textarea>`,
	}
	
	var wikiSource string
	var found bool
	
	for _, pattern := range patterns {
		textareaRegex := regexp.MustCompile(`(?s)` + pattern)
		matches := textareaRegex.FindStringSubmatch(content)
		if len(matches) >= 2 {
			wikiSource = matches[1]
			found = true
			break
		}
	}
	
	if !found {
		return "", fmt.Errorf("could not extract Wikipedia source from edit page")
	}

	// Decode HTML entities
	wikiSource = strings.ReplaceAll(wikiSource, "&lt;", "<")
	wikiSource = strings.ReplaceAll(wikiSource, "&gt;", ">")
	wikiSource = strings.ReplaceAll(wikiSource, "&quot;", "\"")
	wikiSource = strings.ReplaceAll(wikiSource, "&#34;", "\"")
	wikiSource = strings.ReplaceAll(wikiSource, "&apos;", "'")
	wikiSource = strings.ReplaceAll(wikiSource, "&#39;", "'")
	wikiSource = strings.ReplaceAll(wikiSource, "&amp;", "&") // Decode &amp; last
	
	return wikiSource, nil
}

// parseWikitableFromSource extracts and parses wikitable from source
func parseWikitableFromSource(source string) ([]SimpleTable, error) {
	fmt.Printf("Parsing Wikipedia source for tables...\n")
	fmt.Printf("Source length: %d characters\n", len(source))
	
	// More robust table pattern matching
	tablePatterns := []string{
		`(?s)\{\|.*?class=".*?wikitable.*?\n\|\}`,
		`(?s)\{\|.*?wikitable.*?\n\|\}`,
		`(?s)\{\|[^}]*class="[^"]*wikitable[^"]*".*?\|\}`,
	}
	
	var allTables []SimpleTable
	tableMatches := make(map[string]bool) // To avoid duplicates
	
	for _, pattern := range tablePatterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllString(source, -1)
		fmt.Printf("Found %d tables with pattern: %s\n", len(matches), pattern)
		
		for _, match := range matches {
			// Create a hash to avoid duplicates
			hash := fmt.Sprintf("%d", len(match))
			if tableMatches[hash] {
				continue
			}
			tableMatches[hash] = true
			
			table := parseSimpleWikitableContent(match)
			if len(table.Headers) > 0 {
				allTables = append(allTables, table)
			}
		}
	}
	
	fmt.Printf("Total found %d unique tables in Wikipedia source\n", len(allTables))
	
	return allTables, nil
}

// parseSimpleWikitableContent parses a single wikitable to simple format
func parseSimpleWikitableContent(tableSource string) SimpleTable {
	var table SimpleTable
	
	lines := strings.Split(tableSource, "\n")
	var isInHeader bool
	var currentRow []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and table markup
		if line == "" || strings.HasPrefix(line, "{|") || line == "|}" {
			continue
		}
		
		// Header row
		if strings.HasPrefix(line, "!") {
			if len(currentRow) > 0 {
				// Finish previous row
				if len(table.Headers) == 0 {
					table.Headers = currentRow
				} else {
					table.Rows = append(table.Rows, currentRow)
				}
				currentRow = nil
			}
			
			isInHeader = true
			// Split headers by !! or !
			headerText := strings.TrimPrefix(line, "!")
			headers := regexp.MustCompile(`\s*!!\s*|\s*!\s*`).Split(headerText, -1)
			
			for _, header := range headers {
				header = cleanWikiText(header)
				if header != "" {
					currentRow = append(currentRow, header)
				}
			}
		} else if strings.HasPrefix(line, "|-") {
			// Row separator
			if len(currentRow) > 0 {
				if isInHeader && len(table.Headers) == 0 {
					table.Headers = currentRow
				} else {
					table.Rows = append(table.Rows, currentRow)
				}
				currentRow = nil
			}
			isInHeader = false
		} else if strings.HasPrefix(line, "|") && !strings.HasPrefix(line, "|+") {
			// Data row
			if len(currentRow) > 0 && isInHeader && len(table.Headers) == 0 {
				table.Headers = currentRow
				currentRow = nil
			}
			isInHeader = false
			
			// Split cells by || but be smarter about it
			cellText := strings.TrimPrefix(line, "|")
			
			var cells []string
			if strings.Contains(cellText, "||") {
				cells = strings.Split(cellText, "||")
			} else {
				cells = []string{cellText}
			}
			
			for _, cell := range cells {
				cell = strings.TrimSpace(cell)
				if cell != "" {
					cleanedCell := cleanWikiText(cell)
					currentRow = append(currentRow, cleanedCell)
				} else {
					currentRow = append(currentRow, "")
				}
			}
		}
	}
	
	// Add final row
	if len(currentRow) > 0 {
		if len(table.Headers) == 0 {
			table.Headers = currentRow
		} else {
			table.Rows = append(table.Rows, currentRow)
		}
	}
	
	return table
}

// cleanWikiText removes wiki markup from text
func cleanWikiText(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	
	// Handle cell content that starts with HTML attributes
	if strings.Contains(text, "|") && (strings.Contains(text, "style=") || strings.Contains(text, "align=") || strings.Contains(text, "class=")) {
		parts := strings.Split(text, "|")
		if len(parts) > 1 {
			text = parts[len(parts)-1]
		}
	}
	
	// Remove file/image links completely
	text = regexp.MustCompile(`\[\[(?:File|Image):[^\]]*\]\]`).ReplaceAllString(text, "")
	
	// Remove category links completely
	text = regexp.MustCompile(`\[\[Category:[^\]]*\]\]`).ReplaceAllString(text, "")
	
	// Handle piped links [[target|display]] -> keep only display text
	text = regexp.MustCompile(`\[\[[^|\]]*\|([^\]]+)\]\]`).ReplaceAllString(text, "$1")
	
	// Handle simple links [[target]] -> keep only target text
	text = regexp.MustCompile(`\[\[([^\]]+)\]\]`).ReplaceAllString(text, "$1")
	
	// Remove ref tags
	text = regexp.MustCompile(`(?s)<ref[^>]*>.*?</ref>`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`<ref[^>]*/>`).ReplaceAllString(text, "")
	
	// Handle date templates
	dtsRegex := regexp.MustCompile(`\{\{Dts\|(\d{4})\|(\d{1,2})\|(\d{1,2})\}\}`)
	text = dtsRegex.ReplaceAllString(text, "$1-$2-$3")
	
	dtsShortRegex := regexp.MustCompile(`\{\{Dts\|(\d{4})\}\}`)
	text = dtsShortRegex.ReplaceAllString(text, "$1")
	
	// Remove templates
	for i := 0; i < 10; i++ {
		oldText := text
		text = regexp.MustCompile(`\{\{[^{}]*\}\}`).ReplaceAllString(text, "")
		if text == oldText {
			break
		}
	}
	
	// Remove remaining template brackets
	text = strings.ReplaceAll(text, "{{", "")
	text = strings.ReplaceAll(text, "}}", "")
	
	// Remove HTML attributes and markup
	text = regexp.MustCompile(`style\s*=\s*[^|]*\|?`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`class\s*=\s*[^|]*\|?`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`align\s*=\s*[^|]*\|?`).ReplaceAllString(text, "")
	
	// Remove wiki formatting
	text = regexp.MustCompile(`'''([^']+)'''`).ReplaceAllString(text, "$1") // Bold
	text = regexp.MustCompile(`''([^']+)''`).ReplaceAllString(text, "$1")   // Italic
	
	// Clean up whitespace
	text = regexp.MustCompile(`\s*\|\s*`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	
	return text
}

// selectBestTable selects the most suitable table for FCPXML generation
func selectBestTable(tables []SimpleTable) *SimpleTable {
	if len(tables) == 0 {
		return nil
	}
	
	fmt.Printf("Found %d tables, selecting the best one for FCPXML generation\n", len(tables))
	
	bestTable := &tables[0]
	bestScore := 0
	
	for i, table := range tables {
		// Score based on number of headers and data richness
		score := len(table.Headers)
		
		// Bonus for tables with meaningful data
		if len(table.Rows) > 5 {
			score += 5
		}
		if len(table.Rows) > 20 {
			score += 10
		}
		
		// Bonus for tables with date/year columns
		for _, header := range table.Headers {
			headerLower := strings.ToLower(header)
			if strings.Contains(headerLower, "date") || 
			   strings.Contains(headerLower, "year") ||
			   regexp.MustCompile(`^\d{4}$`).MatchString(header) {
				score += 5
			}
		}
		
		// Penalty for single-column tables
		if len(table.Headers) == 1 {
			score -= 10
		}
		
		fmt.Printf("Table %d: %d headers, %d rows, score: %d\n", i+1, len(table.Headers), len(table.Rows), score)
		
		if score > bestScore {
			bestScore = score
			bestTable = &tables[i]
		}
	}
	
	return bestTable
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}