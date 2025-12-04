import os

MAX_IMAGE_SIZE = 10 * 1024 * 1024  # 10 MB
MAX_VIDEO_SIZE = 100 * 1024 * 1024  # 100 MB

# Allowed file types
ALLOWED_IMAGE_EXTENSIONS = {'.jpg', '.jpeg', '.png', '.bmp', '.webp'}
ALLOWED_VIDEO_EXTENSIONS = {'.mp4', '.mov', '.avi', '.mkv', '.flv', '.wmv'}

FRAME_INTERVAL = 5
INSIGHTFACE_SIM_THRESHOLD = 0.35
DEEPFACE_SIM_THRESHOLD = 0.4
VERIFICATION_THRESHOLD = 0.70  
SPOOF_RATE_THRESHOLD = 0.30  

USE_GPU = True
GPU_ID = 0
USE_FP16 = True
BATCH_SIZE = 32
INSIGHTFACE_DET_SIZE = (640, 640)
LIVENESS_MODEL_NAME = 'antispoofing_2.7'
LIVENESS_SCORE_THRESHOLD = 0.9

USE_TENSORRT = True
MAX_WORKSPACE_SIZE = 16 * (1024 ** 3) 
NUM_STREAMS = 16
CUDA_GPU_MEM_LIMIT = 16 * 1024 * 1024 * 1024  # 16 GB
CUDA_ARENA_EXTEND_STRATEGY = 'kSameAsRequested'

FLASK_ENV = os.getenv('FLASK_ENV', 'development')
DEBUG_MODE = os.getenv('DEBUG', 'false').lower() == 'true'
PORT = int(os.getenv('PORT', '5100'))
HOST = os.getenv('HOST', '0.0.0.0')
LOG_LEVEL = os.getenv('LOG_LEVEL', 'INFO')

