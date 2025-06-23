package fcp

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerateConversation tests the conversation generator with a real messages file
func TestGenerateConversation(t *testing.T) {
	tempDir := t.TempDir()

	// Create test messages file (matching messages.txt format)
	messagesFile := filepath.Join(tempDir, "test_messages.txt")
	messagesContent := `Hey u there?
Yes, I'm here.
Good let's chat.
Ok.`

	err := os.WriteFile(messagesFile, []byte(messagesContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test messages file: %v", err)
	}

	// Use actual image files from the Pictures directory
	phoneBackground := "/Users/aa/Pictures/phone_blank001.png"
	blueSpeech := "/Users/aa/Pictures/blue_speech001.png"
	whiteSpeech := "/Users/aa/Pictures/white_speech001.png"
	outputFile := filepath.Join(tempDir, "test_conversation.fcpxml")

	// Test the conversation generator
	err = GenerateConversation(phoneBackground, blueSpeech, whiteSpeech, messagesFile, outputFile)
	if err != nil {
		t.Fatalf("GenerateConversation failed: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output FCPXML file was not created")
	}

	// Read and parse the generated FCPXML
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var fcpxml FCPXML
	err = xml.Unmarshal(data, &fcpxml)
	if err != nil {
		t.Fatalf("Failed to parse generated FCPXML: %v", err)
	}

	// Verify basic structure
	if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 || len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
		t.Fatal("Missing basic FCPXML structure")
	}

	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	// Test 1: Verify phone background video exists
	if len(sequence.Spine.Videos) == 0 {
		t.Fatal("Expected phone background video in spine")
	}

	phoneVideo := &sequence.Spine.Videos[0]
	if !strings.Contains(phoneVideo.Name, "phone_blank001") {
		t.Errorf("Expected phone background video, got: %s", phoneVideo.Name)
	}

	// Test 2: Verify conversation duration (4 messages * 2 seconds + 2 padding = 10 seconds)
	expectedDurationFrames := parseFCPDuration(ConvertSecondsToFCPDuration(10.0))
	actualDurationFrames := parseFCPDuration(phoneVideo.Duration)
	if actualDurationFrames != expectedDurationFrames {
		t.Errorf("Expected phone background duration %d frames, got %d frames", expectedDurationFrames, actualDurationFrames)
	}

	// Test 3: Verify speech bubbles (4 messages = 4 connected videos)
	if len(phoneVideo.NestedVideos) != 4 {
		t.Errorf("Expected 4 connected speech bubble videos, got %d", len(phoneVideo.NestedVideos))
	}

	// Test 4: Verify speech bubble alternation (blue, white, blue, white)
	expectedBubbles := []string{"blue_speech001", "white_speech001", "blue_speech001", "white_speech001"}
	for i, video := range phoneVideo.NestedVideos {
		if !strings.Contains(video.Name, expectedBubbles[i]) {
			t.Errorf("Expected speech bubble %s at index %d, got %s", expectedBubbles[i], i, video.Name)
		}
	}

	// Test 5: Verify speech bubble timing (0s, 2s, 4s, 6s)
	for i, video := range phoneVideo.NestedVideos {
		expectedOffsetSeconds := float64(i * 2)
		expectedOffsetFrames := parseFCPDuration(ConvertSecondsToFCPDuration(expectedOffsetSeconds))
		actualOffsetFrames := parseFCPDuration(video.Offset)
		
		if actualOffsetFrames != expectedOffsetFrames {
			t.Errorf("Expected speech bubble %d offset %d frames, got %d frames", i, expectedOffsetFrames, actualOffsetFrames)
		}
	}

	// Test 6: Verify speech bubble positioning (blue vs white)
	for i, video := range phoneVideo.NestedVideos {
		if i%2 == 0 {
			// Blue bubbles: position="1.26755 -21.1954" scale="0.617236 0.617236" lane="1"
			if video.AdjustTransform == nil || video.AdjustTransform.Position != "1.26755 -21.1954" {
				t.Errorf("Expected blue bubble position '1.26755 -21.1954', got %v", getPositionFromTransform(video.AdjustTransform))
			}
			if video.Lane != "1" {
				t.Errorf("Expected blue bubble lane '1', got '%s'", video.Lane)
			}
		} else {
			// White bubbles: position="0.635834 4.00864" scale="0.653172 0.653172" lane="2"
			if video.AdjustTransform == nil || video.AdjustTransform.Position != "0.635834 4.00864" {
				t.Errorf("Expected white bubble position '0.635834 4.00864', got %v", getPositionFromTransform(video.AdjustTransform))
			}
			if video.Lane != "2" {
				t.Errorf("Expected white bubble lane '2', got '%s'", video.Lane)
			}
		}
	}

	// Test 7: Verify text overlays (4 messages = 4 connected titles)
	if len(phoneVideo.NestedTitles) != 4 {
		t.Errorf("Expected 4 connected text titles, got %d", len(phoneVideo.NestedTitles))
	}

	// Test 8: Verify text content matches messages
	expectedMessages := []string{"Hey u there?", "Yes, I'm here.", "Good let's chat.", "Ok."}
	for i, title := range phoneVideo.NestedTitles {
		actualText := getTextFromTitle(title)
		if actualText != expectedMessages[i] {
			t.Errorf("Expected text '%s' at index %d, got '%s'", expectedMessages[i], i, actualText)
		}
	}

	// Test 9: Verify text positioning (blue vs white messages)
	for i, title := range phoneVideo.NestedTitles {
		actualPosition := getPositionFromTitle(title)
		if i%2 == 0 {
			// Blue messages: Position "0 -3071" lane="4"
			if actualPosition != "0 -3071" {
				t.Errorf("Expected blue text position '0 -3071', got '%s'", actualPosition)
			}
			if title.Lane != "4" {
				t.Errorf("Expected blue text lane '4', got '%s'", title.Lane)
			}
		} else {
			// White messages: Position "0 -1807" lane="3"
			if actualPosition != "0 -1807" {
				t.Errorf("Expected white text position '0 -1807', got '%s'", actualPosition)
			}
			if title.Lane != "3" {
				t.Errorf("Expected white text lane '3', got '%s'", title.Lane)
			}
		}
	}

	// Test 10: Verify text colors (black for blue bubbles, white for white bubbles)
	for i, title := range phoneVideo.NestedTitles {
		if len(title.TextStyleDefs) > 0 {
			actualColor := title.TextStyleDefs[0].TextStyle.FontColor
			if i%2 == 0 {
				// Blue bubble should have black text
				if actualColor != "0 0 0 1" {
					t.Errorf("Expected blue bubble text color '0 0 0 1', got '%s'", actualColor)
				}
			} else {
				// White bubble should have white text
				if actualColor != "0.999995 1 1 1" {
					t.Errorf("Expected white bubble text color '0.999995 1 1 1', got '%s'", actualColor)
				}
			}
		}
	}

	// Test 11: Verify text timing matches speech bubble timing
	for i, title := range phoneVideo.NestedTitles {
		expectedOffsetSeconds := float64(i * 2)
		expectedOffsetFrames := parseFCPDuration(ConvertSecondsToFCPDuration(expectedOffsetSeconds))
		actualOffsetFrames := parseFCPDuration(title.Offset)
		
		if actualOffsetFrames != expectedOffsetFrames {
			t.Errorf("Expected text %d offset %d frames, got %d frames", i, expectedOffsetFrames, actualOffsetFrames)
		}
	}

	// Test 12: Verify XML contains all expected elements
	xmlString := string(data)
	if !strings.Contains(xmlString, "Hey u there?") {
		t.Error("Expected first message not found in XML")
	}
	if !strings.Contains(xmlString, "Yes, I&#39;m here.") {
		t.Error("Expected second message not found in XML")
	}
	if !strings.Contains(xmlString, "adjust-transform") {
		t.Error("Expected adjust-transform elements not found in XML")
	}
}

// TestGenerateConversationErrorCases tests error handling
func TestGenerateConversationErrorCases(t *testing.T) {
	tempDir := t.TempDir()

	// Test 1: Non-existent messages file
	err := GenerateConversation("/test/phone.png", "/test/blue.png", "/test/white.png", "/non/existent.txt", "/test/output.fcpxml")
	if err == nil {
		t.Error("Expected error for non-existent messages file")
	}

	// Test 2: Empty messages file
	emptyFile := filepath.Join(tempDir, "empty.txt")
	err = os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "output.fcpxml")
	err = GenerateConversation("/test/phone.png", "/test/blue.png", "/test/white.png", emptyFile, outputFile)
	if err == nil || !strings.Contains(err.Error(), "no messages found") {
		t.Error("Expected error for empty messages file")
	}
}

// TestReadMessagesFromFile tests the message parsing function
func TestReadMessagesFromFile(t *testing.T) {
	tempDir := t.TempDir()

	// Test with various message formats
	testFile := filepath.Join(tempDir, "test_messages.txt")
	testContent := `First message
Second message

Third message
    Fourth message with spaces    
Fifth message`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	messages, err := readMessagesFromFile(testFile)
	if err != nil {
		t.Fatalf("readMessagesFromFile failed: %v", err)
	}

	expectedMessages := []string{"First message", "Second message", "Third message", "Fourth message with spaces", "Fifth message"}
	if len(messages) != len(expectedMessages) {
		t.Errorf("Expected %d messages, got %d", len(expectedMessages), len(messages))
	}

	for i, expected := range expectedMessages {
		if i < len(messages) && messages[i] != expected {
			t.Errorf("Expected message %d: '%s', got '%s'", i, expected, messages[i])
		}
	}
}

// Helper functions for testing
func getPositionFromTransform(transform *AdjustTransform) string {
	if transform != nil {
		return transform.Position
	}
	return ""
}

func getTextFromTitle(title Title) string {
	if title.Text != nil && len(title.Text.TextStyles) > 0 {
		return title.Text.TextStyles[0].Text
	}
	return ""
}

func getPositionFromTitle(title Title) string {
	for _, param := range title.Params {
		if param.Name == "Position" {
			return param.Value
		}
	}
	return ""
}