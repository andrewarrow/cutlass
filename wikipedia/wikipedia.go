package wikipedia

import (
	"cutlass/fcp"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	
	"golang.org/x/net/html"
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
	// Fetch Wikipedia HTML
	fmt.Printf("Fetching Wikipedia page for: %s\n", articleTitle)
	doc, err := fetchWikipediaHTML(articleTitle)
	if err != nil {
		return fmt.Errorf("failed to fetch Wikipedia page: %v", err)
	}

	// Parse the HTML to extract tables
	fmt.Printf("Parsing Wikipedia HTML for tables...\n")
	tables, err := parseWikitablesFromHTML(doc)
	if err != nil {
		return fmt.Errorf("failed to parse Wikipedia tables: %v", err)
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

// fetchWikipediaHTML fetches the rendered HTML of a Wikipedia article
func fetchWikipediaHTML(articleTitle string) (*html.Node, error) {
	encodedTitle := url.QueryEscape(articleTitle)
	pageURL := fmt.Sprintf("https://en.wikipedia.org/wiki/%s", encodedTitle)
	
	fmt.Printf("Fetching Wikipedia page from: %s\n", pageURL)
	
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Wikipedia page: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}
	
	// Parse the HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}
	
	return doc, nil
}

// parseWikitablesFromHTML extracts and parses wikitable from HTML document
func parseWikitablesFromHTML(doc *html.Node) ([]SimpleTable, error) {
	fmt.Printf("Parsing Wikipedia HTML for tables...\n")
	
	var allTables []SimpleTable
	
	// Find all table elements with class="wikitable"
	tables := findTableElements(doc)
	fmt.Printf("Found %d wikitable elements\n", len(tables))
	
	for i, tableNode := range tables {
		table := parseHTMLTable(tableNode)
		if len(table.Headers) > 0 && len(table.Rows) > 0 {
			allTables = append(allTables, table)
			fmt.Printf("Table %d: %d headers, %d rows\n", i+1, len(table.Headers), len(table.Rows))
		}
	}
	
	fmt.Printf("Total found %d valid tables in Wikipedia HTML\n", len(allTables))
	return allTables, nil
}

// findTableElements finds all HTML table elements with class="wikitable"
func findTableElements(n *html.Node) []*html.Node {
	var tables []*html.Node
	
	if n.Type == html.ElementNode && n.Data == "table" {
		// Check if this table has class="wikitable"
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "wikitable") {
				tables = append(tables, n)
				break
			}
		}
	}
	
	// Recursively search child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		tables = append(tables, findTableElements(c)...)
	}
	
	return tables
}

// parseHTMLTable parses an HTML table element into SimpleTable format
func parseHTMLTable(tableNode *html.Node) SimpleTable {
	var table SimpleTable
	var headerFound bool
	
	// Find all rows (tr elements)
	rows := findElementsByTag(tableNode, "tr")
	
	for _, row := range rows {
		// Check if this row contains header cells (th elements)
		headers := findElementsByTag(row, "th")
		if len(headers) > 0 && !headerFound {
			// This is a header row
			for _, header := range headers {
				headerText := extractTextContent(header)
				if headerText != "" {
					table.Headers = append(table.Headers, headerText)
				}
			}
			headerFound = true
		} else {
			// This is a data row - look for td elements
			cells := findElementsByTag(row, "td")
			if len(cells) > 0 {
				var rowData []string
				for _, cell := range cells {
					cellText := extractTextContent(cell)
					rowData = append(rowData, cellText)
				}
				if len(rowData) > 0 {
					table.Rows = append(table.Rows, rowData)
				}
			}
		}
	}
	
	return table
}

// findElementsByTag finds all child elements with the given tag name
func findElementsByTag(n *html.Node, tag string) []*html.Node {
	var elements []*html.Node
	
	if n.Type == html.ElementNode && n.Data == tag {
		elements = append(elements, n)
	}
	
	// Recursively search child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		elements = append(elements, findElementsByTag(c, tag)...)
	}
	
	return elements
}

// extractTextContent extracts clean text content from an HTML node
func extractTextContent(n *html.Node) string {
	var text strings.Builder
	
	if n.Type == html.TextNode {
		text.WriteString(n.Data)
	}
	
	// Recursively get text from child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(extractTextContent(c))
	}
	
	// Clean up the text
	result := text.String()
	
	// Remove extra whitespace and newlines
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)
	
	// Remove common unwanted content
	result = strings.ReplaceAll(result, "[edit]", "")
	result = strings.ReplaceAll(result, "edit", "")
	
	return result
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