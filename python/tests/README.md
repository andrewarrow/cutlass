# FCPXML Python Library Test Suite

This test suite ensures the FCPXML Python library generates crash-free Final Cut Pro XML files by testing all critical patterns identified from the working Go implementation.

## Test Structure

### 🚨 Crash Prevention Tests (`test_crash_prevention.py`)
Tests the most critical patterns that prevent Final Cut Pro crashes:
- **Smart Collections**: Ensures all 5 required smart collections are generated
- **Video Asset Properties**: Verifies videos never have `hasAudio`/`audioSources` properties
- **Image Asset Properties**: Ensures images have `duration="0s"` and no `frameDuration`
- **Library Structure**: Validates required library location and version
- **File Path Handling**: Tests absolute path conversion

### 🎬 Timeline Element Tests (`test_timeline_elements.py`)
Tests proper timeline element creation and separation:
- **Image Elements**: Images create `<video>` elements with `start` attribute
- **Video Elements**: Videos create `<asset-clip>` elements without `start` attribute
- **Mixed Media**: Tests combinations of images and videos
- **Timeline Ordering**: Verifies elements are ordered by offset
- **Duration Calculation**: Tests timeline duration calculations

### 📋 XML Structure Tests (`test_xml_structure.py`)
Tests XML structure compliance and validation:
- **FCPXML Root**: Version 1.13 and proper attributes
- **Smart Collections**: All 5 collections with correct match rules
- **Resource Structure**: Assets, formats, and resource IDs
- **Well-formed XML**: Validates generated XML syntax
- **UID Uniqueness**: Ensures all UIDs are unique within document

### 🔍 Media Detection Tests (`test_media_detection.py`)
Tests media file property detection and handling:
- **Video Properties**: ffprobe integration and fallback defaults
- **File Type Detection**: Correct handling of images vs videos
- **Audio Detection**: Proper audio track detection
- **Frame Rate Parsing**: Various frame rate format handling
- **Path Conversion**: Relative to absolute path conversion

### 🔗 Integration Tests (`test_integration.py`)
End-to-end tests for complete FCPXML generation:
- **Full Workflow**: Media files to final FCPXML
- **Large Collections**: Performance with many files
- **Mixed Media**: Images and videos together
- **Validation**: XML well-formedness and FCP compatibility
- **Edge Cases**: Empty media lists and error handling

## Running Tests

### Quick Test Run
```bash
python -m pytest tests/ -v
```

### Using Test Runner
```bash
python run_tests.py
```

### Specific Test Categories
```bash
# Only crash prevention tests
python -m pytest tests/test_crash_prevention.py -v

# Only integration tests  
python -m pytest tests/test_integration.py -v

# Only timeline element tests
python -m pytest tests/test_timeline_elements.py -v
```

## Test Coverage

The test suite covers all critical patterns that were identified by comparing working Go FCPXML with crashing Python FCPXML:

- ✅ **Smart Collections** (5 required collections)
- ✅ **Audio Property Handling** (no audio props on video assets)
- ✅ **Element Type Separation** (images→video, videos→asset-clip)
- ✅ **Format Properties** (frameDuration rules)
- ✅ **Library Structure** (location, version, events)
- ✅ **Media Detection** (ffprobe integration)
- ✅ **Timeline Construction** (ordering, durations, offsets)
- ✅ **XML Validation** (well-formed, schema-compliant)

## Adding New Tests

When adding new tests:

1. **Follow naming convention**: `test_*.py` files, `Test*` classes, `test_*` methods
2. **Use fixtures**: Leverage `conftest.py` fixtures for common setup
3. **Test crash patterns**: Focus on patterns that prevent FCP crashes
4. **Include cleanup**: Always cleanup temporary files
5. **Document purpose**: Clear docstrings explaining what's being tested

## Continuous Testing

Run tests frequently during development to catch regressions:

```bash
# Watch for changes and re-run tests
python -m pytest tests/ --tb=short -x
```

The test suite is designed to catch any changes that might reintroduce the FCP crash patterns we've eliminated.