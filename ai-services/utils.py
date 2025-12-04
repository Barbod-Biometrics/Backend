import os
import logging
from pathlib import Path
from config import ALLOWED_IMAGE_EXTENSIONS, ALLOWED_VIDEO_EXTENSIONS, MAX_IMAGE_SIZE, MAX_VIDEO_SIZE

logger = logging.getLogger(__name__)


class ValidationError(Exception):
    """Custom exception for validation errors"""
    def __init__(self, error_code, message):
        self.error_code = error_code
        self.message = message
        super().__init__(message)


def validate_image_file(file_obj, filename):
    """
    Validate uploaded image file.
    
    Returns: (temp_path, error_dict or None)
    """
    if not file_obj or not filename:
        return None, {
            'error': 'invalid_request',
            'message': 'Image file is required'
        }
    
    # Check extension
    ext = Path(filename).suffix.lower()
    if ext not in ALLOWED_IMAGE_EXTENSIONS:
        return None, {
            'error': 'invalid_file_type',
            'message': f'Invalid image type. Allowed: {", ".join(ALLOWED_IMAGE_EXTENSIONS)}'
        }
    
    # Check size
    file_obj.seek(0, os.SEEK_END)
    size = file_obj.tell()
    file_obj.seek(0)
    
    if size > MAX_IMAGE_SIZE:
        return None, {
            'error': 'file_too_large',
            'message': f'Image too large. Max: {MAX_IMAGE_SIZE / (1024*1024):.0f}MB'
        }
    
    if size == 0:
        return None, {
            'error': 'empty_file',
            'message': 'Image file is empty'
        }
    
    return None, None  # Valid


def validate_video_file(file_obj, filename):
    """
    Validate uploaded video file.
    
    Returns: (temp_path, error_dict or None)
    """
    if not file_obj or not filename:
        return None, {
            'error': 'invalid_request',
            'message': 'Video file is required'
        }
    
    # Check extension
    ext = Path(filename).suffix.lower()
    if ext not in ALLOWED_VIDEO_EXTENSIONS:
        return None, {
            'error': 'invalid_file_type',
            'message': f'Invalid video type. Allowed: {", ".join(ALLOWED_VIDEO_EXTENSIONS)}'
        }
    
    # Check size
    file_obj.seek(0, os.SEEK_END)
    size = file_obj.tell()
    file_obj.seek(0)
    
    if size > MAX_VIDEO_SIZE:
        return None, {
            'error': 'file_too_large',
            'message': f'Video too large. Max: {MAX_VIDEO_SIZE / (1024*1024):.0f}MB'
        }
    
    if size == 0:
        return None, {
            'error': 'empty_file',
            'message': 'Video file is empty'
        }
    
    return None, None  # Valid


def safe_delete_file(filepath):
    """Safely delete a file, ignoring errors"""
    try:
        if filepath and os.path.exists(filepath):
            os.unlink(filepath)
            logger.debug(f"Deleted temp file: {filepath}")
    except Exception as e:
        logger.warning(f"Failed to delete temp file {filepath}: {e}")


def format_error_response(error_code, message, http_code=400):
    """
    Format a standard error response.
    """
    return {
        'success': False,
        'error': error_code,
        'message': message
    }, http_code


def format_success_response(data):
    """
    Format a standard success response.
    """
    return {
        'success': True,
        'data': data
    }, 200
