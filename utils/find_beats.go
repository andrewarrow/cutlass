package utils

import (
	"cutlass/fcp"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// BeatDetection represents a detected dramatic change in the audio
type BeatDetection struct {
	Timestamp float64 // Time in seconds
	Intensity float64 // Relative intensity of the change (0.0 to 1.0)
	Type      string  // "amplitude", "spectral", or "combined"
}

// WAVHeader represents the structure of a WAV file header
type WAVHeader struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte
	Subchunk2Size uint32
}

// AudioAnalyzer handles beat detection in WAV files
type AudioAnalyzer struct {
	sampleRate    uint32
	channels      uint16
	bitsPerSample uint16
	samples       []float64
}

// TestWAVHeader just for debugging WAV file structure
func TestWAVHeader(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Read first 44 bytes as hex
	header := make([]byte, 44)
	n, err := file.Read(header)
	if err != nil {
		fmt.Printf("Error reading header: %v\n", err)
		return
	}

	fmt.Printf("Read %d bytes of header:\n", n)
	for i := 0; i < n; i += 4 {
		fmt.Printf("%08x: ", i)
		for j := 0; j < 4 && i+j < len(header); j++ {
			fmt.Printf("%02x ", header[i+j])
		}
		fmt.Printf(" | ")
		for j := 0; j < 4 && i+j < len(header); j++ {
			c := header[i+j]
			if c >= 32 && c <= 126 {
				fmt.Printf("%c", c)
			} else {
				fmt.Printf(".")
			}
		}
		fmt.Printf("\n")
	}

	// Parse key fields
	if len(header) >= 44 {
		riff := string(header[0:4])
		fileSize := binary.LittleEndian.Uint32(header[4:8])
		wave := string(header[8:12])
		fmt1 := string(header[12:16])
		fmtSize := binary.LittleEndian.Uint32(header[16:20])
		audioFormat := binary.LittleEndian.Uint16(header[20:22])
		channels := binary.LittleEndian.Uint16(header[22:24])
		sampleRate := binary.LittleEndian.Uint32(header[24:28])
		byteRate := binary.LittleEndian.Uint32(header[28:32])
		blockAlign := binary.LittleEndian.Uint16(header[32:34])
		bitsPerSample := binary.LittleEndian.Uint16(header[34:36])
		dataID := string(header[36:40])
		dataSize := binary.LittleEndian.Uint32(header[40:44])

		fmt.Printf("\nParsed WAV info:\n")
		fmt.Printf("RIFF: %s\n", riff)
		fmt.Printf("File size: %d\n", fileSize)
		fmt.Printf("WAVE: %s\n", wave)
		fmt.Printf("fmt: %s\n", fmt1)
		fmt.Printf("fmt size: %d\n", fmtSize)
		fmt.Printf("Audio format: %d\n", audioFormat)
		fmt.Printf("Channels: %d\n", channels)
		fmt.Printf("Sample rate: %d\n", sampleRate)
		fmt.Printf("Byte rate: %d\n", byteRate)
		fmt.Printf("Block align: %d\n", blockAlign)
		fmt.Printf("Bits per sample: %d\n", bitsPerSample)
		fmt.Printf("Data ID: %s\n", dataID)
		fmt.Printf("Data size: %d\n", dataSize)

		totalSamples := dataSize / uint32(blockAlign)
		duration := float64(totalSamples) / float64(sampleRate)
		fmt.Printf("Total samples: %d\n", totalSamples)
		fmt.Printf("Duration: %.2f seconds\n", duration)
	}
}

// HandleFindBeatsCommand processes the find-beats command
func HandleFindBeatsCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: find-beats <file.wav> [output.fcpxml]")
	}

	wavFile := args[0]
	outputFile := "beats.fcpxml"
	if len(args) > 1 {
		outputFile = args[1]
	}
	
	// Analyze audio file
	
	analyzer, err := NewAudioAnalyzer(wavFile)
	if err != nil {
		return fmt.Errorf("failed to load audio file: %v", err)
	}

	beats := analyzer.DetectBeats()
	
	// Output results
	if len(beats) == 0 {
		fmt.Println("No dramatic changes detected in the audio file.")
		return nil
	}

	fmt.Printf("Detected %d dramatic changes in %s:\n\n", len(beats), wavFile)
	for i, beat := range beats {
		fmt.Printf("%d. Time: %.3fs | Intensity: %.2f | Type: %s\n", 
			i+1, beat.Timestamp, beat.Intensity, beat.Type)
	}

	// Generate FCPXML with alternating colors
	err = GenerateBeatsVisualization(wavFile, beats, outputFile)
	if err != nil {
		return fmt.Errorf("failed to generate FCPXML: %v", err)
	}

	fmt.Printf("Generated %s with %d color changes\n", outputFile, len(beats))
	return nil
}

// NewAudioAnalyzer creates a new audio analyzer from a WAV file
func NewAudioAnalyzer(filename string) (*AudioAnalyzer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read RIFF header
	var riffID [4]byte
	var fileSize uint32
	var format [4]byte
	
	binary.Read(file, binary.LittleEndian, &riffID)
	binary.Read(file, binary.LittleEndian, &fileSize)
	binary.Read(file, binary.LittleEndian, &format)
	
	if string(riffID[:]) != "RIFF" || string(format[:]) != "WAVE" {
		return nil, fmt.Errorf("not a valid WAV file")
	}

	var audioFormat, channels, blockAlign, bitsPerSample uint16
	var sampleRate, byteRate uint32
	var dataSize uint32
	
	// Read chunks until we find fmt and data
	foundFmt := false
	foundData := false
	
	for !foundFmt || !foundData {
		var chunkID [4]byte
		var chunkSize uint32
		
		err = binary.Read(file, binary.LittleEndian, &chunkID)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk ID: %v", err)
		}
		
		err = binary.Read(file, binary.LittleEndian, &chunkSize)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk size: %v", err)
		}
		
		chunkIDStr := string(chunkID[:])
		
		if chunkIDStr == "fmt " {
			// Read format chunk
			binary.Read(file, binary.LittleEndian, &audioFormat)
			binary.Read(file, binary.LittleEndian, &channels)
			binary.Read(file, binary.LittleEndian, &sampleRate)
			binary.Read(file, binary.LittleEndian, &byteRate)
			binary.Read(file, binary.LittleEndian, &blockAlign)
			binary.Read(file, binary.LittleEndian, &bitsPerSample)
			
			// Skip any remaining fmt chunk data
			if chunkSize > 16 {
				file.Seek(int64(chunkSize-16), io.SeekCurrent)
			}
			foundFmt = true
			
		} else if chunkIDStr == "data" {
			dataSize = chunkSize
			foundData = true
			break
			
		} else {
			// Skip this chunk
			file.Seek(int64(chunkSize), io.SeekCurrent)
		}
	}

	// Support 16-bit and 32-bit PCM, and 32-bit float
	if audioFormat != 1 && audioFormat != 3 {
		return nil, fmt.Errorf("unsupported audio format: %d (only PCM and IEEE float supported)", audioFormat)
	}
	if bitsPerSample != 16 && bitsPerSample != 32 {
		return nil, fmt.Errorf("unsupported bit depth: %d (only 16-bit and 32-bit supported)", bitsPerSample)
	}

	analyzer := &AudioAnalyzer{
		sampleRate:    sampleRate,
		channels:      channels,
		bitsPerSample: bitsPerSample,
	}

	// Read audio data
	audioData := make([]byte, dataSize)
	_, err = io.ReadFull(file, audioData)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %v", err)
	}

	// Convert to float64 samples based on bit depth and format
	if bitsPerSample == 16 {
		// 16-bit PCM
		analyzer.samples = make([]float64, len(audioData)/2)
		for i := 0; i < len(audioData); i += 2 {
			sample := int16(binary.LittleEndian.Uint16(audioData[i:]))
			analyzer.samples[i/2] = float64(sample) / 32768.0
		}
	} else if bitsPerSample == 32 {
		if audioFormat == 1 {
			// 32-bit PCM
			analyzer.samples = make([]float64, len(audioData)/4)
			for i := 0; i < len(audioData); i += 4 {
				sample := int32(binary.LittleEndian.Uint32(audioData[i:]))
				analyzer.samples[i/4] = float64(sample) / 2147483648.0
			}
		} else {
			// 32-bit IEEE float
			analyzer.samples = make([]float64, len(audioData)/4)
			for i := 0; i < len(audioData); i += 4 {
				bits := binary.LittleEndian.Uint32(audioData[i:])
				sample := math.Float32frombits(bits)
				analyzer.samples[i/4] = float64(sample)
			}
		}
	}

	// If stereo, convert to mono by averaging channels
	if analyzer.channels == 2 {
		monoSamples := make([]float64, len(analyzer.samples)/2)
		for i := 0; i < len(monoSamples); i++ {
			monoSamples[i] = (analyzer.samples[i*2] + analyzer.samples[i*2+1]) / 2.0
		}
		analyzer.samples = monoSamples
	}

	// Audio info for reference
	duration := float64(len(analyzer.samples))/float64(analyzer.sampleRate)
	fmt.Printf("Audio: %.1fs, %d Hz, %d channels\n", duration, analyzer.sampleRate, analyzer.channels)

	return analyzer, nil
}

// DetectBeats finds dramatic changes in the audio using multiple methods
func (a *AudioAnalyzer) DetectBeats() []BeatDetection {
	var allBeats []BeatDetection

	// Method 1: Sudden amplitude changes (like forceful piano chords)
	amplitudeBeats := a.detectAmplitudeChanges()
	allBeats = append(allBeats, amplitudeBeats...)

	// Method 2: Spectral flux for musical transitions
	spectralBeats := a.detectSpectralChanges()
	allBeats = append(allBeats, spectralBeats...)

	// Sort by timestamp and merge nearby detections
	sort.Slice(allBeats, func(i, j int) bool {
		return allBeats[i].Timestamp < allBeats[j].Timestamp
	})

	// Merge nearby beats (within 0.1 seconds)
	mergedBeats := a.mergeNearbyBeats(allBeats, 0.1)

	// Filter out weak detections and return top candidates
	return a.filterStrongBeats(mergedBeats, 0.2)
}

// detectAmplitudeChanges finds sudden increases in audio amplitude
func (a *AudioAnalyzer) detectAmplitudeChanges() []BeatDetection {
	var beats []BeatDetection
	
	// Window size for analysis (100ms)
	windowSize := int(float64(a.sampleRate) * 0.1)
	hopSize := windowSize / 4 // 25ms hops
	
	if windowSize >= len(a.samples) {
		return beats
	}

	var energies []float64
	
	// Calculate RMS energy for each window
	for i := 0; i < len(a.samples)-windowSize; i += hopSize {
		var energy float64
		for j := i; j < i+windowSize; j++ {
			energy += a.samples[j] * a.samples[j]
		}
		energies = append(energies, math.Sqrt(energy/float64(windowSize)))
	}

	// Find dramatic increases in energy
	for i := 1; i < len(energies); i++ {
		prev := energies[i-1]
		curr := energies[i]
		
		// Look for sudden increases (2x or more)
		if curr > prev*2.0 && curr > 0.05 {
			timestamp := float64(i*hopSize) / float64(a.sampleRate)
			intensity := math.Min(curr/0.5, 1.0) // Normalize to 0-1
			
			beats = append(beats, BeatDetection{
				Timestamp: timestamp,
				Intensity: intensity,
				Type:      "amplitude",
			})
		}
	}

	return beats
}

// detectSpectralChanges finds changes in frequency content
func (a *AudioAnalyzer) detectSpectralChanges() []BeatDetection {
	var beats []BeatDetection
	
	// Window size for FFT analysis (200ms)
	windowSize := int(float64(a.sampleRate) * 0.2)
	hopSize := windowSize / 2 // 100ms hops
	
	if windowSize >= len(a.samples) {
		return beats
	}

	var spectralFlux []float64
	var prevMagnitudes []float64

	// Calculate spectral flux
	for i := 0; i < len(a.samples)-windowSize; i += hopSize {
		window := a.samples[i : i+windowSize]
		magnitudes := a.calculateSpectralMagnitudes(window)
		
		if len(prevMagnitudes) > 0 {
			flux := a.calculateSpectralFlux(prevMagnitudes, magnitudes)
			spectralFlux = append(spectralFlux, flux)
		}
		prevMagnitudes = magnitudes
	}

	// Find peaks in spectral flux
	threshold := a.calculateAdaptiveThreshold(spectralFlux, 1.5)
	
	for i := 1; i < len(spectralFlux)-1; i++ {
		if spectralFlux[i] > threshold && 
		   spectralFlux[i] > spectralFlux[i-1] && 
		   spectralFlux[i] > spectralFlux[i+1] {
			
			timestamp := float64((i+1)*hopSize) / float64(a.sampleRate)
			intensity := math.Min(spectralFlux[i]/threshold, 1.0)
			
			beats = append(beats, BeatDetection{
				Timestamp: timestamp,
				Intensity: intensity,
				Type:      "spectral",
			})
		}
	}

	return beats
}

// calculateSpectralMagnitudes performs a simple magnitude spectrum calculation
func (a *AudioAnalyzer) calculateSpectralMagnitudes(window []float64) []float64 {
	// Simple approach: divide into frequency bands and calculate energy
	numBands := 20
	bandSize := len(window) / numBands
	magnitudes := make([]float64, numBands)
	
	for band := 0; band < numBands; band++ {
		start := band * bandSize
		end := start + bandSize
		if end > len(window) {
			end = len(window)
		}
		
		var energy float64
		for i := start; i < end; i++ {
			energy += window[i] * window[i]
		}
		magnitudes[band] = math.Sqrt(energy)
	}
	
	return magnitudes
}

// calculateSpectralFlux measures the change between two magnitude spectra
func (a *AudioAnalyzer) calculateSpectralFlux(prev, curr []float64) float64 {
	var flux float64
	for i := 0; i < len(prev) && i < len(curr); i++ {
		diff := curr[i] - prev[i]
		if diff > 0 {
			flux += diff * diff
		}
	}
	return math.Sqrt(flux)
}

// calculateAdaptiveThreshold calculates a dynamic threshold based on recent values
func (a *AudioAnalyzer) calculateAdaptiveThreshold(values []float64, multiplier float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	return mean * multiplier
}

// mergeNearbyBeats combines beats that are very close in time
func (a *AudioAnalyzer) mergeNearbyBeats(beats []BeatDetection, threshold float64) []BeatDetection {
	if len(beats) == 0 {
		return beats
	}
	
	var merged []BeatDetection
	current := beats[0]
	
	for i := 1; i < len(beats); i++ {
		if beats[i].Timestamp-current.Timestamp <= threshold {
			// Merge: keep the one with higher intensity
			if beats[i].Intensity > current.Intensity {
				current = beats[i]
			}
		} else {
			merged = append(merged, current)
			current = beats[i]
		}
	}
	merged = append(merged, current)
	
	return merged
}

// filterStrongBeats keeps only the most significant beat detections
func (a *AudioAnalyzer) filterStrongBeats(beats []BeatDetection, minIntensity float64) []BeatDetection {
	var filtered []BeatDetection
	
	for _, beat := range beats {
		if beat.Intensity >= minIntensity {
			filtered = append(filtered, beat)
		}
	}
	
	return filtered
}

// GenerateBeatsVisualization creates an FCPXML with alternating background colors and audio
func GenerateBeatsVisualization(wavFile string, beats []BeatDetection, outputFile string) error {
	// Get absolute path for WAV file
	absWavPath, err := filepath.Abs(wavFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for audio: %v", err)
	}

	// Check if audio file exists
	if _, err := os.Stat(absWavPath); os.IsNotExist(err) {
		return fmt.Errorf("audio file does not exist: %s", absWavPath)
	}

	// Calculate total duration from last beat + some extra time
	totalDuration := 10.0 // Default minimum
	if len(beats) > 0 {
		lastBeat := beats[len(beats)-1]
		totalDuration = lastBeat.Timestamp + 5.0 // Add 5 seconds after last beat
	}

	// Create base FCPXML
	fcpxml, err := fcp.GenerateEmpty("")
	if err != nil {
		return fmt.Errorf("failed to create base FCPXML: %v", err)
	}

	// Use proper resource management
	registry := fcp.NewResourceRegistry(fcpxml)
	tx := fcp.NewTransaction(registry)
	defer tx.Rollback()

	// Reserve IDs for audio asset, color generators
	numIDs := 4 // audio asset, audio format, blue generator, green generator
	ids := tx.ReserveIDs(numIDs)
	
	audioAssetID := ids[0]
	audioFormatID := ids[1]
	blueGeneratorID := ids[2]
	greenGeneratorID := ids[3]
	
	// Create audio asset and format
	audioName := filepath.Base(absWavPath)
	audioName = audioName[:len(audioName)-len(filepath.Ext(audioName))] // Remove extension
	
	// For now, use the total duration we calculated
	audioDuration := fcp.ConvertSecondsToFCPDuration(totalDuration)
	
	_, err = tx.CreateFormatWithFrameDuration(audioFormatID, "1001/24000s", "", "", "")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create audio format: %v", err)
	}
	
	audioAsset, err := tx.CreateAsset(audioAssetID, absWavPath, audioName, audioDuration, audioFormatID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create audio asset: %v", err)
	}

	// Create Vivid generator effects for blue and green
	_, err = tx.CreateEffect(blueGeneratorID, "Vivid Blue", ".../Generators.localized/Solids.localized/Vivid.localized/Vivid.motn")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create blue generator: %v", err)
	}
	_, err = tx.CreateEffect(greenGeneratorID, "Vivid Green", ".../Generators.localized/Solids.localized/Vivid.localized/Vivid.motn")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create green generator: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Get the sequence to add color changes
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	
	// Create color segments based on beats
	currentTime := 0.0
	isBlue := true // Start with blue

	for _, beat := range beats {
		// Create color segment from currentTime to beat.Timestamp
		segmentDuration := beat.Timestamp - currentTime
		
		if segmentDuration > 0 {
			var generatorID string
			var colorName string
			
			if isBlue {
				generatorID = blueGeneratorID
				colorName = "Blue Background"
			} else {
				generatorID = greenGeneratorID
				colorName = "Green Background"
			}

			// Create video element for this color segment
			colorVideo := fcp.Video{
				Ref:      generatorID,
				Offset:   fcp.ConvertSecondsToFCPDuration(currentTime),
				Name:     colorName,
				Duration: fcp.ConvertSecondsToFCPDuration(segmentDuration),
				Start:    "86486400/24000s", // Use standard start time like samples
				Lane:     "1", // Put on connected storyline above audio
			}

			// Remove custom color parameters for now - use default generator colors

			sequence.Spine.Videos = append(sequence.Spine.Videos, colorVideo)
		}

		currentTime = beat.Timestamp
		isBlue = !isBlue // Alternate color
	}

	// Add final color segment from last beat to end
	if currentTime < totalDuration {
		finalDuration := totalDuration - currentTime
		
		var generatorID string
		var colorName string
		
		if isBlue {
			generatorID = blueGeneratorID
			colorName = "Blue Background"
		} else {
			generatorID = greenGeneratorID
			colorName = "Green Background"
		}

		finalVideo := fcp.Video{
			Ref:      generatorID,
			Offset:   fcp.ConvertSecondsToFCPDuration(currentTime),
			Name:     colorName,
			Duration: fcp.ConvertSecondsToFCPDuration(finalDuration),
			Start:    "86486400/24000s",
			Lane:     "1", // Put on connected storyline above audio
		}

		// Use default generator colors

		sequence.Spine.Videos = append(sequence.Spine.Videos, finalVideo)
	}

	// Add audio asset-clip to spine as primary storyline
	audioClip := fcp.AssetClip{
		Ref:      audioAsset.ID,
		Offset:   "0s",
		Name:     audioAsset.Name,
		Duration: audioDuration,
		Start:    "0s",
		Format:   audioAsset.Format,
		TCFormat: "NDF",
	}
	
	sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, audioClip)

	// Update sequence duration
	sequence.Duration = fcp.ConvertSecondsToFCPDuration(totalDuration)

	// Write the FCPXML file
	return fcp.WriteToFile(fcpxml, outputFile)
}