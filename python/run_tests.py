#!/usr/bin/env python3
"""
Test runner for FCPXML Python Library.

This script runs the complete test suite and provides a summary of results.
"""

import subprocess
import sys
from pathlib import Path


def run_tests():
    """Run the complete test suite."""
    print("🧪 Running FCPXML Python Library Test Suite")
    print("=" * 50)
    
    # Change to the project directory
    project_dir = Path(__file__).parent
    
    try:
        # Run pytest with verbose output
        result = subprocess.run([
            sys.executable, "-m", "pytest", 
            "tests/", 
            "-v", 
            "--tb=short",
            "--color=yes"
        ], cwd=project_dir, check=False, capture_output=False)
        
        print("\n" + "=" * 50)
        
        if result.returncode == 0:
            print("✅ All tests passed!")
            print("\n🎯 Key areas covered:")
            print("   • FCP crash prevention patterns")
            print("   • Smart collections generation")
            print("   • Video vs image element handling")
            print("   • Media file detection and properties")
            print("   • XML structure validation")
            print("   • End-to-end FCPXML generation")
        else:
            print("❌ Some tests failed!")
            print("   Check the output above for details.")
            
        return result.returncode == 0
        
    except FileNotFoundError:
        print("❌ pytest not found. Install with: pip install pytest")
        return False
    except Exception as e:
        print(f"❌ Error running tests: {e}")
        return False


if __name__ == "__main__":
    success = run_tests()
    sys.exit(0 if success else 1)