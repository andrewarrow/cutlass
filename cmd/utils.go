package cmd

import (
	"cutlass/creative"
	"cutlass/utils"

	"github.com/spf13/cobra"
)

var utilsCmd = &cobra.Command{
	Use:   "utils",
	Short: "Utility commands",
	Long:  "Miscellaneous utility commands for various tasks.",
}

var genaudioCmd = &cobra.Command{
	Use:   "genaudio <file.txt> [voice]",
	Short: "Generate audio files from simple text file (one sentence per line)",
	Long: `Generate audio files from a simple text file format.

The input file should have one sentence per line. Empty lines are skipped.
Uses the filename (without extension) as the video ID.

Optional voice parameter: If provided, all lines will use the specified voice.
Available voices: agucchi, algernon, amanda, archibald, australian, china, deep, doug, drew, dundee, elsa, hank, harry, heather, iran, jane, jessica, karen, kevin, kosovo, mike, miss, mrs, pepe, peter, rachel, richie, saltburn, sara, steve, tommy, vatra, yoav

Example with waymo.txt:
- Creates ./data/waymo_audio/ directory
- Generates s1_duration.wav, s2_duration.wav, etc.
- Duration is automatically detected and added to filename

Examples:
cutlass utils genaudio waymo.txt
cutlass utils genaudio waymo.txt David

Audio files are generated using chatterbox TTS and skip existing files.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.HandleGenAudioCommand(args)
		return nil
	},
}

var parseVttCmd = &cobra.Command{
	Use:   "parse-vtt <file.vtt>",
	Short: "Parse VTT file and extract plain text content",
	Long: `Parse a WebVTT subtitle file and extract only the plain text content.

Removes all timing information, formatting tags, and positioning data.
Outputs clean text suitable for further processing.

Example:
cutlass utils parse-vtt data/video.en.vtt`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.HandleParseVttCommand(args)
		return nil
	},
}

var parseVttAndCutCmd = &cobra.Command{
	Use:   "parse-vtt-and-cut <video-id>",
	Short: "Parse VTT file and cut video into sentence clips",
	Long: `Parse a WebVTT subtitle file and use timecodes to cut video into individual sentence clips.

Reads ./data/<video-id>.en.vtt for subtitle timecodes and text.
Cuts ./data/<video-id>.mov into clips stored in ./data/<video-id>/ directory.
Each clip is named with sentence number and duration.

Example:
cutlass utils parse-vtt-and-cut iPSP_j-QyX4

This will:
- Read ./data/iPSP_j-QyX4.en.vtt for timecodes
- Cut ./data/iPSP_j-QyX4.mov into clips
- Store clips in ./data/iPSP_j-QyX4/ directory`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.HandleParseVttAndCutCommand(args)
		return nil
	},
}

var genaudioPlayCmd = &cobra.Command{
	Use:   "genaudio-play <play.json>",
	Short: "Generate audio files from play JSON with consistent character voices",
	Long: `Generate audio files from a play JSON file using chatterbox TTS with consistent character voice mapping.

The input JSON should follow the play format with dialogue entries:
{
  "act": "II",
  "scene": "2", 
  "title": "Scene Title",
  "setting": "Scene description",
  "dialogue": [
    {
      "character": "Character Name",
      "stage_direction": "speaking slowly",
      "line": "The dialogue text to speak"
    }
  ]
}

Features:
- Consistent voice assignment per character using MD5 hash
- Automatic numbering (001.wav, 002.wav, etc.)
- Skips existing audio files to avoid regeneration
- Uses chatterbox utah.py for high-quality TTS
- Creates ./data/{basename}_audio/ directory structure

Example:
cutlass utils genaudio-play play.json`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.HandleGenAudioPlayCommand(args)
		return nil
	},
}

var creativeTextCmd = &cobra.Command{
	Use:   "creative-text <input.json> [output.fcpxml]",
	Short: "Generate creative animated text presentation from JSON",
	Long: `Generate a creative animated text presentation from a JSON file containing sections and points.

The input JSON should be an array of objects with "title" and "points" fields:
[
  {
    "title": "Section Title",
    "points": ["Point 1", "Point 2", "Point 3"]
  }
]

Features:
- 🚀 EXPLOSIVE text animations with dramatic slide-ins and overshoot effects
- 💥 Each section CRASHES in from different directions with scale animations
- ⚡ Bullet points fly in with increasing intensity - each one BIGGER than the last!
- 🎯 All caps text with emojis for maximum visual impact  
- 🎬 Perfect for picture-in-picture video - guaranteed to captivate your audience
- 🌟 Professional timing with dramatic pauses for maximum effect

Example:
cutlass utils creative-text jenny_hansen_lane.json output.fcpxml`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return creative.HandleCreativeTextCommand(args)
	},
}

var addShadowTextCmd = &cobra.Command{
	Use:   "add-shadow-text <file.txt> [output.fcpxml] [duration-seconds]",
	Short: "Generate shadow text FCPXML from text file",
	Long: `Generate FCPXML with shadow text effects from a simple text file.

The input file should contain the text content. The text will be broken into
small chunks (sometimes just 1 word, sometimes 2 small words) and placed as
text elements with shadow formatting on the timeline.

Features:
- Automatic text chunking with intelligent word grouping
- Customizable total duration (defaults to auto-calculated based on text length)
- Text elements are distributed evenly across the specified duration
- Adaptive font sizing (600px for short text, 460px for longer text)
- Professional shadow text styling matching samples/shadow_text.fcpxml
- Creative text splitting for visual impact preserving word spacing
- Proper FCPXML structure with frame-aligned timing

Font and Style:
- Font: Avenir Next Condensed Heavy Italic
- Colors: Bright magenta text with yellow shadow
- Shadow offset: 26x317 with 20px blur radius
- Center alignment with custom kerning

Example:
cutlass utils add-shadow-text shadow.txt
cutlass utils add-shadow-text shadow.txt custom_output.fcpxml
cutlass utils add-shadow-text shadow.txt output.fcpxml 30`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.HandleAddShadowTextCommand(args)
		return nil
	},
}

var fxStaticImageCmd = &cobra.Command{
	Use:   "fx-static-image <image.png|image1.png,image2.png> [output.fcpxml] [effect-type]",
	Short: "Generate dynamic FCPXML with sophisticated multi-phase animated effects for static PNG images",
	Long: `Generate FCPXML with sophisticated multi-phase animation effects to transform static PNG images into cinematic video content.

This command creates dramatic video movement using advanced Final Cut Pro techniques:

🎬 ENHANCED MULTI-PHASE CAMERA MOVEMENT:
- Phase 1 (0-25%): SLOW gentle drift and zoom-in with smooth easing
- Phase 2 (25-50%): FAST panning and rotation with quick transitions  
- Phase 3 (50-75%): SUPER FAST dramatic movement with maximum speed
- Phase 4 (75-100%): SLOW elegant settle with smooth finish
- Variable speed timing creates cinematic tension and release

🎯 SOPHISTICATED ANIMATION STACK:
- Position: Complex multi-directional movement (-80 to +80 pixels)
- Scale: Dynamic zoom cycles (100% → 140% → 90% → 160% → 125%)
- Rotation: Dramatic tilt changes (-4° to +4° with speed variations)
- Anchor Point: Dynamic pivot points for interesting rotation centers
- Professional easing curves (linear, easeIn, easeOut, smooth)

⚡ BUILT-IN FINAL CUT PRO EFFECTS:
- Shape Mask (FFSuperEllipseMask) for subtle 3D perspective
- Handheld camera shake simulation
- Rounded corners and depth effects
- Professional parameter settings from working samples

🔧 TECHNICAL FEATURES:
- Frame-aligned timing with 23.976fps timebase compliance
- 5 keyframes per animation parameter for smooth motion
- Proper resource management with transaction system
- FCP-compatible effect UIDs verified from samples
- DTD-compliant FCPXML structure

Effect Types:
Standard: shake, perspective, flip, 360-tilt, 360-pan, light-rays, glow, cinematic (default)
Creative: parallax, breathe, pendulum, elastic, spiral, figure8, heartbeat, wind
Advanced: inner-collapse (digital mind breakdown with complex multi-layer animation)
Cinematic: shatter-archive (nostalgic stop-motion with analog photography decay)
Special: 
  potpourri (cycles through all effects at 1-second intervals)
  variety-pack (random effect per image, perfect for multiple images)

Examples:
Single image:
cutlass utils fx-static-image photo.png
cutlass utils fx-static-image photo.png dynamic_video.fcpxml
cutlass utils fx-static-image photo.png output.fcpxml heartbeat
cutlass utils fx-static-image photo.png parallax

Multiple images (10 seconds each):
cutlass utils fx-static-image image1.png,image2.png,image3.png output.fcpxml potpourri
cutlass utils fx-static-image photo1.png,photo2.png heartbeat
cutlass utils fx-static-image a.png,b.png,c.png,d.png variety-pack

Word-bounce with custom colors and duration:
cutlass utils fx-static-image image.png word-bounce -c blue -o red -d 20
WORDS='hello,world,test' cutlass utils fx-static-image image.png word-bounce -c green -o black -d 15`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fontColor, _ := cmd.Flags().GetString("font-color")
		outlineColor, _ := cmd.Flags().GetString("outline-color")
		duration, _ := cmd.Flags().GetFloat64("duration")
		utils.HandleFXStaticImageCommandWithColorAndDuration(args, fontColor, outlineColor, duration)
		return nil
	},
}

var findBeatsCmd = &cobra.Command{
	Use:   "find-beats <file.wav>",
	Short: "Detect dramatic musical changes and beat points in WAV audio files",
	Long: `Analyze WAV audio files to detect dramatic changes in music such as forceful piano chords,
sudden dynamics shifts, and other musical transitions.

This command uses advanced audio analysis techniques to identify moments where the music
has dramatic changes - perfect for timing video cuts, transitions, and effects.

Detection Methods:
- Amplitude Analysis: Detects sudden increases in volume/energy (like forceful piano chords)
- Spectral Analysis: Identifies changes in frequency content (musical transitions)
- Combined Analysis: Merges nearby detections for accurate timing

The algorithm specifically looks for:
🎹 Forceful piano chord strikes and dramatic volume changes
🎵 Musical transitions between sections
🥁 Percussive hits and rhythmic emphasis points
🎶 Build-ups followed by dramatic drops or changes

Output Format:
Each detected change shows:
- Timestamp in seconds (precise to milliseconds)
- Intensity score (0.0 to 1.0, higher = more dramatic)
- Detection type (amplitude, spectral, or combined)

Requirements:
- Input must be 16-bit PCM WAV format
- Mono or stereo supported (stereo converted to mono)
- Works best with music that has clear dynamic changes

Example:
cutlass utils find-beats song.wav

Sample Output:
1. Time: 15.234s | Intensity: 0.87 | Type: amplitude
2. Time: 32.156s | Intensity: 0.92 | Type: spectral
3. Time: 48.789s | Intensity: 0.95 | Type: combined`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return utils.HandleFindBeatsCommand(args)
	},
}

var txtConvoCmd = &cobra.Command{
	Use:   "txt-convo <conversation.txt> [output.fcpxml] [duration-per-message]",
	Short: "Generate iMessage-style conversation animation from text file",
	Long: `Generate FCPXML with iMessage-style conversation animation from a text file.

The input file should contain conversation in "Name: message" format, one per line:

Example input file:
John: Hey, how are you?
Sarah: I'm doing great, thanks for asking!
John: Want to grab coffee later?
Sarah: Sure! What time works for you?
John: How about 3pm at the usual place?
Sarah: Perfect, see you then! 😊

Features:
📱 Authentic iMessage appearance with:
- iPhone background from ./assets/iphone_bg.png
- Blue bubbles for first person (./assets/bubble_blue.png)
- Gray bubbles for other participants (./assets/bubble_gray.png)
- Proper bubble positioning (right/left alignment)

🎬 Smooth animations:
- Messages slide in sequentially like real iMessage
- Bubbles animate from off-screen with scale effect
- Stacked conversation layout with proper spacing
- Customizable timing per message

📝 Text handling:
- Clean, readable text over speech bubbles
- Supports emoji and special characters
- Automatic text positioning over bubbles
- Proper font sizing for mobile readability

Requirements:
- iPhone background: ./assets/iphone_bg.png
- Blue bubble: ./assets/bubble_blue.png  
- Gray bubble: ./assets/bubble_gray.png
- Input text file with "Name: message" format

Examples:
cutlass utils txt-convo conversation.txt
cutlass utils txt-convo chat.txt output.fcpxml 
cutlass utils txt-convo messages.txt final.fcpxml 3.0

Duration per message defaults to 2.5 seconds for natural pacing.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.HandleTxtConvoCommand(args)
		return nil
	},
}

func init() {
	utilsCmd.AddCommand(genaudioCmd)
	utilsCmd.AddCommand(genaudioPlayCmd)
	utilsCmd.AddCommand(parseVttCmd)
	utilsCmd.AddCommand(parseVttAndCutCmd)
	utilsCmd.AddCommand(creativeTextCmd)
	utilsCmd.AddCommand(addShadowTextCmd)
	utilsCmd.AddCommand(fxStaticImageCmd)
	utilsCmd.AddCommand(findBeatsCmd)
	utilsCmd.AddCommand(txtConvoCmd)
	
	// Add flags for fx-static-image command
	fxStaticImageCmd.Flags().StringP("font-color", "c", "pink", "Font color as English name (red, blue, green, yellow, etc.) or RGBA values (0-1 format)")
	fxStaticImageCmd.Flags().StringP("outline-color", "o", "black", "Outline color as English name (red, blue, green, yellow, etc.) or RGBA values (0-1 format)")
	fxStaticImageCmd.Flags().Float64P("duration", "d", 9.0, "Duration in seconds for word-bounce effect (default: 9.0)")
}
