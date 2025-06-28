"""
XML validation utilities.
"""

import subprocess


def run_xml_validation(xml_file_path: str) -> tuple[bool, str]:
    """
    Run basic XML well-formedness validation using xmllint.
    
    🚨 CRITICAL: XML must be well-formed for FCPXML (from schema.yaml)
    Note: Full DTD validation requires Apple's DTD but basic validation catches most issues.
    """
    try:
        result = subprocess.run(
            ['xmllint', '--noout', xml_file_path],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            return True, ""
        else:
            # Extract meaningful error from xmllint output
            error_msg = result.stderr.strip()
            return False, error_msg
            
    except subprocess.TimeoutExpired:
        return False, "XML validation timed out"
    except subprocess.CalledProcessError as e:
        return False, f"xmllint error: {e.stderr}"
    except FileNotFoundError:
        return False, "xmllint not found - install libxml2-utils"