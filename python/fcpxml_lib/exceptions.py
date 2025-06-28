"""
Exception classes for FCPXML library.
"""


class FCPXMLError(Exception):
    """Base exception for FCPXML generation errors"""
    pass


class ValidationError(FCPXMLError):
    """Raised when validation rules are violated"""
    pass