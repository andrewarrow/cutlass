# UID Generation Logic for `./cutlass youtube-bulk-assemble`

## Summary
**All UIDs are hardcoded** - there is no dynamic UID generation in the current implementation.

## Command Flow
When running `./cutlass youtube-bulk-assemble 5ids.txt "test name"`:

1. **main.go:32** → calls `youtube.HandleYouTubeBulkAssembleCommand(args)`
2. **youtube/bulk.go:57** → `HandleYouTubeBulkAssembleCommand()` function:
   - Reads video IDs from the file (e.g., `5ids.txt`)
   - Calls `AssembleTop5FCPXML()` with IDs and name
   - Outputs to `{name}_top5.fcpxml` (spaces replaced with underscores)
3. **youtube/bulk.go** → `AssembleTop5FCPXML()` calls `fcp.GenerateTop5FCPXML()`
4. **fcp/generator.go** → `GenerateTop5FCPXML()` processes the template

## UID Implementation Details

### Hardcoded UID Values
All UIDs come from the template file `/Users/aa/cs/cutlass/templates/top5.fcpxml`:

- **Video Assets**: `1060CDAD22BE0C793867C03043293F21`
  - Used for all video files (qtdmugxbjog.mov, i6AHys3pwyc.mov, ECXAFUmdJkI.mov, etc.)
  - Same UID shared across all video assets in the project

- **Logo Images**: `3F909E29CB42ED246BCE2FB34A45EF33`
- **Audio Assets**: `798E3523EBB845A996446F857DE6AA7E` 
- **Sound Effects**: `3DC9B020540A135712966448E4C2AB96`

### Template Processing
The template uses Go's `text/template` syntax:
```xml
<asset id="r{{add $index 10}}" name="{{$video.ID}}" uid="1060CDAD22BE0C793867C03043293F21" ...>
```

- **Dynamic**: `id` (r10, r11, r12, etc.) and `name` (video ID from file)
- **Static**: `uid` value is hardcoded in template

### No UID Generation Logic
The codebase contains **zero** UID generation functionality:
- No UUID libraries imported
- No random number generation for UIDs
- No hash-based ID creation
- No crypto/rand usage
- No unique identifier algorithms

## Current Behavior
When you run the command:
1. All generated video assets get the same UID: `1060CDAD22BE0C793867C03043293F21`
2. This UID is identical across all videos in the same project
3. This UID is identical across different projects generated by this tool
4. Only the `id` attribute and `name` attribute change between assets

## Files Involved
- **main.go**: Command routing
- **youtube/bulk.go**: Command handler and file processing
- **fcp/generator.go**: Template processing and FCPXML generation
- **templates/top5.fcpxml**: Source of all hardcoded UID values

## Implications
- All video assets share the same UID within and across projects
- FCPXML files generated by this tool will have duplicate UIDs
- Final Cut Pro may treat assets with identical UIDs as the same media