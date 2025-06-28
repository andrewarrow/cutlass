"""
Schema loading utilities.
"""

import yaml
from pathlib import Path
from typing import Dict, Any

_SCHEMA = None


def load_schema() -> Dict[str, Any]:
    """Load the FCPXML rules schema from schema.yaml"""
    schema_path = Path(__file__).parent.parent.parent / "schema.yaml"
    with open(schema_path, 'r') as f:
        return yaml.safe_load(f)


def get_schema() -> Dict[str, Any]:
    """Get cached schema or load it if not already loaded"""
    global _SCHEMA
    if _SCHEMA is None:
        _SCHEMA = load_schema()
    return _SCHEMA