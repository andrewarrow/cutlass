
### Installation
```bash
# Clone the future of video generation
git clone https://github.com/your-username/cutlass
cd cutlass
go build

# Start creating magic
./cutlass --help
```

### Quick Start Example

**YouTube Video → Auto-Segmented Clips**  
```bash
./cutlass youtube dQw4w9WgXcQ
./cutlass vtt-clips dQw4w9WgXcQ.en.vtt 00:30_10,01:45_15,02:30_20
# Generates: segmented_video.fcpxml  
# Result: Perfectly timed clips with transition animations
```

## 📚 Learn More: FCPXML Resources

Want to dive deeper into the mysterious world of FCPXML? Here are the resources that built Cutlass's expertise:

### Essential Reading
- [FCP.cafe Developer Resources](https://fcp.cafe/developers/fcpxml/) - Community knowledge base
- [Apple FCPXML Reference](https://developer.apple.com/documentation/professional-video-applications/fcpxml-reference) - Official documentation  
- [CommandPost DTD Files](https://github.com/CommandPost/CommandPost/tree/develop/src/extensions/cp/apple/fcpxml/dtd) - XML validation schemas
- [DAWFileKit](https://github.com/orchetect/DAWFileKit) - Swift FCPXML parser

