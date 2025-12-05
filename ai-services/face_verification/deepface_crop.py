import cv2
import logging
from deepface import DeepFace

logger = logging.getLogger(__name__)


def create_passport_photo(card_image_path, output_path, insightface_app=None):
    try:
        img = cv2.imread(card_image_path)
        if img is None:
            logger.error(f"Could not read image: {card_image_path}")
            return False, {
                'error': 'image_read_failed',
                'message': 'Failed to read image file'
            }
        
        if img.shape[0] < 100 or img.shape[1] < 100:
            logger.error(f"Image too small: {img.shape}")
            return False, {
                'error': 'image_too_small',
                'message': 'Image dimensions too small (minimum 100x100)'
            }
        
        if insightface_app is not None:
            faces = insightface_app.get(img)
            if not faces:
                return False, {
                    'error': 'no_face_detected',
                    'message': 'No face detected in the provided image'
                }
            if len(faces) > 1:
                return False, {
                    'error': 'multiple_faces',
                    'message': f'Multiple faces detected ({len(faces)}). Provide image with exactly one face'
                }
            face = faces[0]
            x, y, w, h = int(face.bbox[0]), int(face.bbox[1]), int(face.bbox[2] - face.bbox[0]), int(face.bbox[3] - face.bbox[1])
        else:
            try:
                faces = DeepFace.extract_faces(img_path=card_image_path, enforce_detection=True)
            except Exception as e:
                logger.error(f"Face detection failed: {e}")
                return False, {
                    'error': 'face_detection_failed',
                    'message': 'No face detected in image or invalid image format'
                }
            
            if not faces:
                logger.warning("No faces detected")
                return False, {
                    'error': 'no_face_detected',
                    'message': 'No face detected in the provided image'
                }
            
            if len(faces) > 1:
                logger.warning(f"Multiple faces detected: {len(faces)}")
                return False, {
                    'error': 'multiple_faces',
                    'message': f'Multiple faces detected ({len(faces)}). Provide image with exactly one face'
                }
            
            # Get face coordinates
            face = faces[0]['facial_area']
            x, y, w, h = face['x'], face['y'], face['w'], face['h']
        
        # Passport aspect ratio (2:3)
        target_ratio = 5/6
        
        # Calculate dimensions - head takes 70% of height (less top padding)
        target_height = int(h / 0.6)
        target_width = int(target_height * target_ratio)
        
        # Center calculation - minimal upward shift for less top padding
        center_x = x + w // 2
        center_y = y + h // 2  # Center on face rather than above it
        
        crop_x = max(0, center_x - target_width // 2)
        crop_y = max(0, center_y - target_height // 2)
        crop_w = min(img.shape[1] - crop_x, target_width)
        crop_h = min(img.shape[0] - crop_y, target_height)
        
        passport_photo = img[crop_y:crop_y+crop_h, crop_x:crop_x+crop_w]
        
        if passport_photo.size == 0:
            logger.error("Cropped image is empty")
            return False, {
                'error': 'crop_failed',
                'message': 'Failed to crop image'
            }
        
        # Save result
        if not cv2.imwrite(output_path, passport_photo):
            logger.error(f"Failed to write output: {output_path}")
            return False, {
                'error': 'write_failed',
                'message': 'Failed to save cropped image'
            }
        
        logger.info(f"Passport photo created: {output_path} ({passport_photo.shape[1]}x{passport_photo.shape[0]})")
        
        return True, {
            'width': passport_photo.shape[1],
            'height': passport_photo.shape[0],
            'aspect_ratio': round(passport_photo.shape[1] / passport_photo.shape[0], 3)
        }
        
    except Exception as e:
        logger.exception(f"Unexpected error in passport photo creation: {e}")
        return False, {
            'error': 'internal_error',
            'message': 'An unexpected error occurred'
        }
