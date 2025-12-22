import logging
import tempfile
import os
import sys
from pathlib import Path

from flask import Flask, request, jsonify, send_file, redirect
from flask_cors import CORS
from flask_swagger_ui import get_swaggerui_blueprint

sys.path.insert(0, os.path.dirname(__file__))

from combined_face_verification import CombinedFaceVerification
from deepface_crop import create_passport_photo
import config
from utils import validate_image_file, validate_video_file, safe_delete_file, format_error_response
from insightface.app import FaceAnalysis
try:
    from insightface.model_zoo import get_model
    HAS_LIVENESS = True
except Exception:
    HAS_LIVENESS = False
    get_model = None

logging.basicConfig(
    level=getattr(logging, config.LOG_LEVEL),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
STATIC_DIR = os.path.join(BASE_DIR, 'static')

app = Flask(__name__, static_folder=STATIC_DIR, static_url_path='/static')
CORS(app)

INSIGHTFACE_APP = None
LIVENESS_MODEL = None
MODELS_INITIALIZED = False


def load_config_from_env():
    import logging
    logger = logging.getLogger(__name__)
    
    env_mappings = {
        'USE_GPU': ('USE_GPU', lambda x: x.lower() in ('true', '1', 'yes')),
        'GPU_ID': ('GPU_ID', int),
        'USE_FP16': ('USE_FP16', lambda x: x.lower() in ('true', '1', 'yes')),
        'BATCH_SIZE': ('BATCH_SIZE', int),
        'USE_TENSORRT': ('USE_TENSORRT', lambda x: x.lower() in ('true', '1', 'yes')),
        'FRAME_INTERVAL': ('FRAME_INTERVAL', int),
        'INSIGHTFACE_SIM_THRESHOLD': ('INSIGHTFACE_SIMILARITY_THRESHOLD', float),
        'DEEPFACE_SIM_THRESHOLD': ('DEEPFACE_SIMILARITY_THRESHOLD', float),
        'VERIFICATION_THRESHOLD': ('VERIFICATION_THRESHOLD', float),
        'SPOOF_RATE_THRESHOLD': ('SPOOF_RATE_THRESHOLD', float),
        'LIVENESS_SCORE_THRESHOLD': ('LIVENESS_SCORE_THRESHOLD', float),
    }
    
    loaded = []
    for config_key, (env_key, converter) in env_mappings.items():
        env_value = os.getenv(env_key)
        if env_value is not None:
            try:
                setattr(config, config_key, converter(env_value))
                loaded.append(f"{config_key}={getattr(config, config_key)}")
            except (ValueError, TypeError) as e:
                logger.warning(f"Failed to load {env_key} from environment: {e}")
    
    if loaded:
        logger.info(f"Loaded config from environment: {', '.join(loaded)}")


def initialize_models():
    global INSIGHTFACE_APP, LIVENESS_MODEL, MODELS_INITIALIZED
    import logging
    logger = logging.getLogger(__name__)
    # runtime check helpers
    import ctypes
    from ctypes.util import find_library

    def _has_tensorrt_lib():
        candidates = [
            'libnvinfer.so.10',
            'libnvinfer.so',
            find_library('nvinfer')
        ]
        for c in candidates:
            if not c:
                continue
            try:
                ctypes.CDLL(c)
                return True
            except Exception:
                continue
        return False

    use_gpu = config.USE_GPU
    try:
        import onnxruntime
        available_providers = onnxruntime.get_available_providers()
        logger.info(f"📋 ONNX Runtime available providers: {available_providers}")

        if use_gpu:
            has_cuda = 'CUDAExecutionProvider' in available_providers
            has_tensorrt_ep = 'TensorrtExecutionProvider' in available_providers
            has_tensorrt_libs = _has_tensorrt_lib()

            if not has_cuda and not has_tensorrt_ep:
                logger.warning("⚠️  GPU requested but no GPU providers available in ONNX Runtime, falling back to CPU")
                use_gpu = False
            elif config.USE_TENSORRT and has_tensorrt_ep and not has_tensorrt_libs:
                logger.warning("⚠️  TensorrtExecutionProvider present but TensorRT system libraries not found (libnvinfer). Disabling TensorRT provider to avoid EP load errors.")
                # We'll avoid adding Tensorrt EP below
                has_tensorrt_ep = False
            else:
                logger.info(f"✅ GPU providers available - CUDA: {has_cuda}, TensorRT EP: {has_tensorrt_ep}, TensorRT libs: {has_tensorrt_libs}")
    except Exception as e:
        logger.warning(f"⚠️  Could not check GPU availability: {e}; falling back to CPU")
        use_gpu = False

    providers = []
    if use_gpu:
        try:
            import onnxruntime
            available_providers = onnxruntime.get_available_providers()
        except Exception:
            available_providers = []

        if 'CUDAExecutionProvider' in available_providers:
            providers.append(('CUDAExecutionProvider', {
                'device_id': config.GPU_ID,
                'gpu_mem_limit': getattr(config, 'CUDA_GPU_MEM_LIMIT', 16 * 1024 * 1024 * 1024),
                'arena_extend_strategy': getattr(config, 'CUDA_ARENA_EXTEND_STRATEGY', 'kSameAsRequested')
            }))

        if config.USE_TENSORRT and 'TensorrtExecutionProvider' in available_providers and _has_tensorrt_lib():
            providers.append(('TensorrtExecutionProvider', {
                'device_id': config.GPU_ID,
                'trt_fp16_enable': config.USE_FP16,
                'trt_max_workspace_size': config.MAX_WORKSPACE_SIZE,
            }))

        if not providers:
            providers = ['CPUExecutionProvider']
            use_gpu = False
    else:
        providers = ['CPUExecutionProvider']

    try:
        logger.info(f"Using ONNX providers: {providers}")
        INSIGHTFACE_APP = FaceAnalysis(providers=providers)
        INSIGHTFACE_APP.prepare(ctx_id=config.GPU_ID if use_gpu else -1, det_size=config.INSIGHTFACE_DET_SIZE)
        try:
            INSIGHTFACE_APP._prepared_on_gpu = bool(use_gpu)
        except Exception:
            pass

        try:
            import numpy as np
            dummy = np.zeros((config.INSIGHTFACE_DET_SIZE[0], config.INSIGHTFACE_DET_SIZE[1], 3), dtype=np.uint8)
            INSIGHTFACE_APP.get(dummy)
        except Exception:
            logger.debug('InsightFace warm-up failed or not applicable')

        LIVENESS_MODEL = None
        if HAS_LIVENESS and get_model is not None:
            try:
                LIVENESS_MODEL = get_model(config.LIVENESS_MODEL_NAME)
                if LIVENESS_MODEL:
                    ctx_id = config.GPU_ID if use_gpu else -1
                    LIVENESS_MODEL.prepare(ctx_id=ctx_id, input_size=(128, 128))
                    try:
                        LIVENESS_MODEL._prepared_on_gpu = bool(use_gpu)
                    except Exception:
                        pass
                    # Warm-up liveness
                    try:
                        import numpy as np
                        LIVENESS_MODEL.get(np.zeros((128, 128, 3), dtype=np.uint8))
                    except Exception:
                        logger.debug('Liveness warm-up failed or not applicable')
            except Exception as e:
                logger.warning(f"⚠️ Could not load liveness model: {e}")

        MODELS_INITIALIZED = True
        logger.info('✅ Models initialized and warmed up')
    except Exception as e:
        logger.exception(f"Model initialization failed: {e}")
        INSIGHTFACE_APP = None
        LIVENESS_MODEL = None
        MODELS_INITIALIZED = False


SWAGGER_URL = '/api/docs'
API_URL = '/swagger.json'
swaggerui_blueprint = get_swaggerui_blueprint(
    SWAGGER_URL,
    API_URL,
    config={'app_name': "Face Verification API"}
)
app.register_blueprint(swaggerui_blueprint, url_prefix=SWAGGER_URL)


@app.route('/swagger')
def swagger_redirect():
    return redirect(SWAGGER_URL)


# Serve a lightweight Swagger UI page (CDN) to avoid filesystem/static routing issues.
@app.route('/api/docs')
@app.route('/api/docs/')
def swagger_ui_page():
        html = '''<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>API Docs - Face Verification</title>
        <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" />
    </head>
    <body>
        <div id="swagger-ui"></div>
        <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
        <script>
            window.onload = function() {
                SwaggerUIBundle({
                    url: '/swagger.json',
                    dom_id: '#swagger-ui',
                    deepLinking: true,
                    presets: [
                        SwaggerUIBundle.presets.apis,
                    ],
                    layout: 'BaseLayout'
                });
            };
        </script>
    </body>
</html>'''
        return html, 200, {'Content-Type': 'text/html'}


# ============================================================================
# Static Files
# ============================================================================

@app.route('/')
def index():
    """Serve the frontend"""
    return send_file(os.path.join(STATIC_DIR, 'index.html'))


@app.route('/swagger.json', methods=['GET'])
def swagger():
    """Generate and serve Swagger/OpenAPI specification dynamically"""
    swagger_spec = {
        "openapi": "3.0.0",
        "info": {
            "title": "Face Verification API",
            "version": "1.0.0",
            "description": "API for face cropping and video-based face verification"
        },
        "servers": [
            {"url": "http://localhost:" + str(config.PORT)}
        ],
        "paths": {
            "/health": {
                "get": {
                    "summary": "Health check",
                    "tags": ["System"],
                    "responses": {
                        "200": {
                            "description": "Service is healthy",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "type": "object",
                                        "properties": {
                                            "status": {"type": "string"},
                                            "service": {"type": "string"},
                                            "version": {"type": "string"}
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            },
            "/crop": {
                "post": {
                    "summary": "Crop image to passport-style photo",
                    "tags": ["Image Processing"],
                    "requestBody": {
                        "required": True,
                        "content": {
                            "multipart/form-data": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "image": {"type": "string", "format": "binary"}
                                    },
                                    "required": ["image"]
                                }
                            }
                        }
                    },
                    "responses": {
                        "200": {
                            "description": "Cropped image",
                            "content": {
                                "image/jpeg": {"schema": {"type": "string", "format": "binary"}}
                            }
                        },
                        "400": {
                            "description": "Validation error",
                            "content": {
                                "application/json": {
                                    "schema": {"$ref": "#/components/schemas/ErrorResponse"}
                                }
                            }
                        }
                    }
                }
            },
            "/verify": {
                "post": {
                    "summary": "Verify face in video matches photo",
                    "tags": ["Face Verification"],
                    "requestBody": {
                        "required": True,
                        "content": {
                            "multipart/form-data": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "photo": {"type": "string", "format": "binary"},
                                        "video": {"type": "string", "format": "binary"}
                                    },
                                    "required": ["photo", "video"]
                                }
                            }
                        }
                    },
                    "responses": {
                        "200": {
                            "description": "Verification result",
                            "content": {
                                "application/json": {
                                    "schema": {"$ref": "#/components/schemas/VerifyResponse"}
                                }
                            }
                        },
                        "400": {
                            "description": "Verification or validation failure",
                            "content": {
                                "application/json": {
                                    "schema": {"$ref": "#/components/schemas/ErrorResponse"}
                                }
                            }
                        }
                    }
                }
            },
            "/info": {
                "get": {
                    "summary": "Service info and limits",
                    "tags": ["System"],
                    "responses": {
                        "200": {
                            "description": "Service info",
                            "content": {
                                "application/json": {
                                    "schema": {"type": "object"}
                                }
                            }
                        }
                    }
                }
            },
            "/config": {
                "get": {
                    "summary": "Get current service configuration",
                    "tags": ["Configuration"],
                    "responses": {
                        "200": {
                            "description": "Current configuration",
                            "content": {
                                "application/json": {
                                    "schema": {"type": "object"}
                                }
                            }
                        }
                    }
                },
                "post": {
                    "summary": "Update service configuration dynamically",
                    "tags": ["Configuration"],
                    "requestBody": {
                        "required": True,
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "object",
                                    "properties": {
                                        "frame_interval": {"type": "integer", "minimum": 1},
                                        "insightface_similarity_threshold": {"type": "number", "minimum": 0, "maximum": 1},
                                        "deepface_similarity_threshold": {"type": "number", "minimum": 0, "maximum": 1},
                                        "verification_threshold": {"type": "number", "minimum": 0, "maximum": 1},
                                        "spoof_rate_threshold": {"type": "number", "minimum": 0, "maximum": 1},
                                        "use_gpu": {"type": "boolean"},
                                        "gpu_id": {"type": "integer", "minimum": 0},
                                        "use_fp16": {"type": "boolean"},
                                        "batch_size": {"type": "integer", "minimum": 1},
                                        "use_tensorrt": {"type": "boolean"},
                                        "liveness_score_threshold": {"type": "number", "minimum": 0, "maximum": 1}
                                    }
                                }
                            }
                        }
                    },
                    "responses": {
                        "200": {
                            "description": "Configuration updated successfully",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "type": "object",
                                        "properties": {
                                            "success": {"type": "boolean"},
                                            "message": {"type": "string"},
                                            "updated_params": {"type": "object"},
                                            "current_config": {"type": "object"}
                                        }
                                    }
                                }
                            }
                        },
                        "400": {
                            "description": "Invalid configuration",
                            "content": {
                                "application/json": {
                                    "schema": {"$ref": "#/components/schemas/ErrorResponse"}
                                }
                            }
                        }
                    }
                }
            }
        },
        "components": {
            "schemas": {
                "ErrorResponse": {
                    "type": "object",
                    "properties": {
                        "error": {"type": "string"},
                        "message": {"type": "string"}
                    }
                },
                "VerifyResponse": {
                    "type": "object",
                    "properties": {
                        "success": {"type": "boolean"},
                        "reason": {"type": "string"},
                        "message": {"type": "string"},
                        "stats": {"type": "object"},
                        "results": {"type": "array"},
                        "messages": {"type": "array"}
                    }
                }
            }
        }
    }
    return jsonify(swagger_spec)


# ============================================================================
# Error Handlers
# ============================================================================

@app.errorhandler(400)
def bad_request(e):
    return format_error_response('bad_request', str(e.description), 400)


@app.errorhandler(404)
def not_found(e):
    return format_error_response('not_found', 'Endpoint not found', 404)


@app.errorhandler(500)
def internal_error(e):
    logger.exception("Internal server error")
    return format_error_response('internal_error', 'Internal server error', 500)


# ============================================================================
# Routes
# ============================================================================

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        'status': 'ok',
        'service': 'face-verification',
        'version': '1.0.0'
    }), 200


@app.route('/crop', methods=['POST'])
def crop():
    """
    Crop image to passport-style photo.
    
    Request: multipart/form-data with 'image' field
    Response: image/jpeg (on success) or JSON error
    """
    logger.info("POST /crop request received")
    
    # Validate input
    if 'image' not in request.files:
        logger.warning("Missing 'image' field in crop request")
        return format_error_response('missing_field', "Provide 'image' file", 400)
    
    file = request.files['image']
    if not file.filename:
        return format_error_response('empty_filename', 'Filename is empty', 400)
    
    # Validate file
    _, err = validate_image_file(file, file.filename)
    if err:
        logger.warning(f"Image validation failed: {err['error']}")
        return format_error_response(err['error'], err['message'], 400)
    
    # Create temp files
    input_tmp = tempfile.NamedTemporaryFile(
        suffix=Path(file.filename).suffix or '.jpg',
        delete=False
    )
    input_path = input_tmp.name
    input_tmp.close()
    
    output_tmp = tempfile.NamedTemporaryFile(suffix='.jpg', delete=False)
    output_path = output_tmp.name
    output_tmp.close()
    
    try:
        file.save(input_path)
        logger.debug(f"Saved input image to {input_path}")
        
        # Crop image
        success, result = create_passport_photo(input_path, output_path, insightface_app=INSIGHTFACE_APP)
        
        if not success:
            logger.warning(f"Crop failed: {result['error']}")
            return format_error_response(result['error'], result['message'], 400)
        
        logger.info(f"Crop successful: {result}")
        
        # Send file
        return send_file(
            output_path,
            mimetype='image/jpeg',
            as_attachment=True,
            download_name='cropped.jpg'
        )
        
    except Exception as e:
        logger.exception(f"Crop endpoint error: {e}")
        return format_error_response('internal_error', 'Failed to process image', 500)
    finally:
        safe_delete_file(input_path)
        # Note: output_path is sent to client, don't delete yet


@app.route('/verify', methods=['POST'])
def verify():
    """
    Verify face in video matches photo.
    
    Request: multipart/form-data with 'photo' and 'video' fields
    Response: JSON with verification result and detailed stats
    """
    logger.info("POST /verify request received")
    
    # Validate inputs
    if 'photo' not in request.files or 'video' not in request.files:
        logger.warning("Missing photo or video field")
        return format_error_response('missing_fields', "Provide 'photo' and 'video' files", 400)
    
    photo_file = request.files['photo']
    video_file = request.files['video']
    
    if not photo_file.filename or not video_file.filename:
        return format_error_response('empty_filename', 'Filename is empty', 400)
    
    # Validate files
    _, photo_err = validate_image_file(photo_file, photo_file.filename)
    if photo_err:
        logger.warning(f"Photo validation failed: {photo_err['error']}")
        return format_error_response(photo_err['error'], photo_err['message'], 400)
    
    _, video_err = validate_video_file(video_file, video_file.filename)
    if video_err:
        logger.warning(f"Video validation failed: {video_err['error']}")
        return format_error_response(video_err['error'], video_err['message'], 400)
    
    # Create temp files
    photo_tmp = tempfile.NamedTemporaryFile(
        suffix=Path(photo_file.filename).suffix or '.jpg',
        delete=False
    )
    photo_path = photo_tmp.name
    photo_tmp.close()
    
    video_tmp = tempfile.NamedTemporaryFile(
        suffix=Path(video_file.filename).suffix or '.mp4',
        delete=False
    )
    video_path = video_tmp.name
    video_tmp.close()
    
    try:
        photo_file.save(photo_path)
        video_file.save(video_path)
        logger.debug(f"Saved photo to {photo_path}, video to {video_path}")
        
        try:
            verifier = CombinedFaceVerification(
                base_image_path=photo_path,
                frame_interval=config.FRAME_INTERVAL,
                insightface_similarity_threshold=config.INSIGHTFACE_SIM_THRESHOLD,
                deepface_similarity_threshold=config.DEEPFACE_SIM_THRESHOLD,
                spoof_threshold=config.SPOOF_RATE_THRESHOLD,
                insightface_app=INSIGHTFACE_APP,
                liveness_model=LIVENESS_MODEL
            )
            logger.info("Verifier initialized")
        except ValueError as e:
            logger.warning(f"Base image error: {e}")
            error_msg = str(e)
            if 'multiple faces' in error_msg.lower():
                error_code = 'multiple_faces_in_photo'
            elif 'no face' in error_msg.lower():
                error_code = 'no_face_in_photo'
            else:
                error_code = 'base_image_error'
            return format_error_response(error_code, error_msg, 400)
        except Exception as e:
            logger.exception(f"Verifier initialization error: {e}")
            return format_error_response('initialization_error', 'Failed to initialize verifier', 500)
        
        result = verifier.process_video(video_path)
        logger.info(f"Verification result: {result['reason']}")
        
        http_code = 200 if result['success'] else 400
        return jsonify(result), http_code
        
    except Exception as e:
        logger.exception(f"Verify endpoint error: {e}")
        return format_error_response('internal_error', 'Verification failed', 500)
    finally:
        safe_delete_file(photo_path)
        safe_delete_file(video_path)


# ============================================================================
# Info Endpoints
# ============================================================================

@app.route('/info', methods=['GET'])
def info():
    """Get service configuration and limits"""
    return jsonify({
        'service': 'Face Verification API',
        'version': '1.0.0',
        'endpoints': {
            'health': 'GET /health',
            'crop': 'POST /crop',
            'verify': 'POST /verify',
            'info': 'GET /info',
            'config': 'GET/POST /config'
        },
        'limits': {
            'max_image_size_mb': config.MAX_IMAGE_SIZE / (1024 * 1024),
            'max_video_size_mb': config.MAX_VIDEO_SIZE / (1024 * 1024),
            'allowed_image_types': list(config.ALLOWED_IMAGE_EXTENSIONS),
            'allowed_video_types': list(config.ALLOWED_VIDEO_EXTENSIONS)
        },
        'thresholds': {
            'insightface_similarity': config.INSIGHTFACE_SIM_THRESHOLD,
            'deepface_similarity': config.DEEPFACE_SIM_THRESHOLD,
            'verification_rate': config.VERIFICATION_THRESHOLD,
            'spoof_rate_max': config.SPOOF_RATE_THRESHOLD
        }
    }), 200


@app.route('/config', methods=['GET'])
def get_config():
    """Get current service configuration"""
    return jsonify({
        'processing': {
            'frame_interval': config.FRAME_INTERVAL,
            'insightface_similarity_threshold': config.INSIGHTFACE_SIM_THRESHOLD,
            'deepface_similarity_threshold': config.DEEPFACE_SIM_THRESHOLD,
            'verification_threshold': config.VERIFICATION_THRESHOLD,
            'spoof_rate_threshold': config.SPOOF_RATE_THRESHOLD
        },
        'hardware': {
            'use_gpu': config.USE_GPU,
            'gpu_id': config.GPU_ID,
            'use_fp16': config.USE_FP16,
            'batch_size': config.BATCH_SIZE,
            'use_tensorrt': config.USE_TENSORRT
        },
        'liveness': {
            'model_name': config.LIVENESS_MODEL_NAME,
            'score_threshold': config.LIVENESS_SCORE_THRESHOLD
        },
        'file_limits': {
            'max_image_size_mb': config.MAX_IMAGE_SIZE / (1024 * 1024),
            'max_video_size_mb': config.MAX_VIDEO_SIZE / (1024 * 1024)
        }
    }), 200


@app.route('/config', methods=['POST'])
def update_config():
    logger.info("POST /config request received")
    
    try:
        data = request.get_json()
        if not data:
            return format_error_response('empty_request', 'Request body cannot be empty', 400)
        
        allowed_params = {
            'frame_interval': int,
            'insightface_similarity_threshold': float,
            'deepface_similarity_threshold': float,
            'verification_threshold': float,
            'spoof_rate_threshold': float,
            'use_gpu': bool,
            'gpu_id': int,
            'use_fp16': bool,
            'batch_size': int,
            'cuda_gpu_mem_limit': int,
            'cuda_arena_extend_strategy': str,
            'use_tensorrt': bool,
            'liveness_score_threshold': float
        }
        
        updated_params = {}
        
        for key, value in data.items():
            if key not in allowed_params:
                logger.warning(f"Attempted to update non-configurable parameter: {key}")
                return format_error_response('invalid_parameter', f"Parameter '{key}' is not configurable", 400)
            
            expected_type = allowed_params[key]
            try:
                if isinstance(value, expected_type):
                    converted_value = value
                else:
                    converted_value = expected_type(value)
                
                if key == 'frame_interval' and converted_value < 1:
                    return format_error_response('invalid_value', 'frame_interval must be >= 1', 400)
                if 'threshold' in key and not (0.0 <= converted_value <= 1.0):
                    return format_error_response('invalid_value', f'{key} must be between 0.0 and 1.0', 400)
                if key == 'batch_size' and converted_value < 1:
                    return format_error_response('invalid_value', 'batch_size must be >= 1', 400)
                if key == 'gpu_id' and converted_value < 0:
                    return format_error_response('invalid_value', 'gpu_id must be >= 0', 400)
                
                setattr(config, key.upper(), converted_value)
                updated_params[key] = converted_value
                logger.info(f"Updated config: {key} = {converted_value}")
                
            except (ValueError, TypeError) as e:
                logger.warning(f"Failed to convert {key} to {expected_type.__name__}: {e}")
                return format_error_response('type_error', f"Invalid type for '{key}': expected {expected_type.__name__}", 400)
        
        return jsonify({
            'success': True,
            'message': f'Updated {len(updated_params)} configuration parameter(s)',
            'updated_params': updated_params,
            'current_config': {
                'processing': {
                    'frame_interval': config.FRAME_INTERVAL,
                    'insightface_similarity_threshold': config.INSIGHTFACE_SIM_THRESHOLD,
                    'deepface_similarity_threshold': config.DEEPFACE_SIM_THRESHOLD,
                    'verification_threshold': config.VERIFICATION_THRESHOLD,
                    'spoof_rate_threshold': config.SPOOF_RATE_THRESHOLD
                },
                'hardware': {
                    'use_gpu': config.USE_GPU,
                    'gpu_id': config.GPU_ID,
                    'use_fp16': config.USE_FP16,
                    'batch_size': config.BATCH_SIZE,
                    'use_tensorrt': config.USE_TENSORRT
                },
                'liveness': {
                    'model_name': config.LIVENESS_MODEL_NAME,
                    'score_threshold': config.LIVENESS_SCORE_THRESHOLD
                }
            }
        }), 200
        
    except Exception as e:
        logger.exception(f"Config update error: {e}")
        return format_error_response('internal_error', 'Failed to update configuration', 500)

    finally:
        try:
            hw_keys = {'use_gpu', 'gpu_id', 'use_fp16', 'use_tensorrt', 'cuda_gpu_mem_limit', 'cuda_arena_extend_strategy'}
            if any(k in updated_params for k in hw_keys):
                logger = logging.getLogger(__name__)
                logger.info('Hardware-related config changed, reinitializing models')
                initialize_models()
        except Exception:
            pass


@app.route('/config/<param>', methods=['PATCH'])
def edit_config_param(param):
    logger.info(f"PATCH /config/{param} request received")
    try:
        data = request.get_json()
        if not data or 'value' not in data:
            return format_error_response('empty_request', "Request body must include 'value'", 400)

        value = data['value']

        allowed_params = {
            'frame_interval': int,
            'insightface_similarity_threshold': float,
            'deepface_similarity_threshold': float,
            'verification_threshold': float,
            'spoof_rate_threshold': float,
            'use_gpu': bool,
            'gpu_id': int,
            'use_fp16': bool,
            'batch_size': int,
            'cuda_gpu_mem_limit': int,
            'cuda_arena_extend_strategy': str,
            'use_tensorrt': bool,
            'liveness_score_threshold': float
        }

        if param not in allowed_params:
            logger.warning(f"Attempted to update non-configurable parameter: {param}")
            return format_error_response('invalid_parameter', f"Parameter '{param}' is not configurable", 400)

        expected_type = allowed_params[param]
        try:
            if isinstance(value, expected_type):
                converted_value = value
            else:
                converted_value = expected_type(value)

            if param == 'frame_interval' and converted_value < 1:
                return format_error_response('invalid_value', 'frame_interval must be >= 1', 400)
            if 'threshold' in param and not (0.0 <= converted_value <= 1.0):
                return format_error_response('invalid_value', f"{param} must be between 0.0 and 1.0", 400)
            if param == 'batch_size' and converted_value < 1:
                return format_error_response('invalid_value', 'batch_size must be >= 1', 400)
            if param == 'gpu_id' and converted_value < 0:
                return format_error_response('invalid_value', 'gpu_id must be >= 0', 400)

            setattr(config, param.upper(), converted_value)
            logger.info(f"Updated config: {param} = {converted_value}")

            if param in {'use_gpu', 'gpu_id', 'use_fp16', 'use_tensorrt', 'cuda_gpu_mem_limit', 'cuda_arena_extend_strategy'}:
                logger.info('Hardware-related config changed via PATCH, reinitializing models')
                try:
                    initialize_models()
                except Exception:
                    logger.exception('Failed to reinitialize models after config change')

            return jsonify({
                'success': True,
                'message': f"Updated '{param}'",
                'updated_param': {param: converted_value},
                'current_config': {
                    'processing': {
                        'frame_interval': config.FRAME_INTERVAL,
                        'insightface_similarity_threshold': config.INSIGHTFACE_SIM_THRESHOLD,
                        'deepface_similarity_threshold': config.DEEPFACE_SIM_THRESHOLD,
                        'verification_threshold': config.VERIFICATION_THRESHOLD,
                        'spoof_rate_threshold': config.SPOOF_RATE_THRESHOLD
                    },
                    'hardware': {
                        'use_gpu': config.USE_GPU,
                        'gpu_id': config.GPU_ID,
                        'use_fp16': config.USE_FP16,
                        'batch_size': config.BATCH_SIZE,
                        'use_tensorrt': config.USE_TENSORRT,
                        'cuda_gpu_mem_limit': config.CUDA_GPU_MEM_LIMIT,
                        'cuda_arena_extend_strategy': config.CUDA_ARENA_EXTEND_STRATEGY
                    },
                    'liveness': {
                        'model_name': config.LIVENESS_MODEL_NAME,
                        'score_threshold': config.LIVENESS_SCORE_THRESHOLD
                    }
                }
            }), 200

        except (ValueError, TypeError) as e:
            logger.warning(f"Failed to convert {param} to {expected_type.__name__}: {e}")
            return format_error_response('type_error', f"Invalid type for '{param}': expected {expected_type.__name__}", 400)

    except Exception as e:
        logger.exception(f"Config param edit error: {e}")
        return format_error_response('internal_error', 'Failed to update configuration parameter', 500)

# ============================================================================
# Main
# ============================================================================

if __name__ == '__main__':
    load_config_from_env()
    
    logger.info(f"Starting Face Verification Service on {config.HOST}:{config.PORT}")
    logger.info(f"Environment: {config.FLASK_ENV}")
    logger.info(f"GPU Mode: {'ENABLED' if config.USE_GPU else 'DISABLED (using CPU)'}")
    logger.info(f"Batch Size: {config.BATCH_SIZE}")
    logger.info(f"FP16: {'ENABLED' if config.USE_FP16 else 'DISABLED'}")
    logger.info(f"TensorRT: {'ENABLED' if config.USE_TENSORRT else 'DISABLED'}")
    try:
        initialize_models()
    except Exception:
        logger.exception('Model initialization at startup failed')

    app.run(
        host=config.HOST,
        port=config.PORT,
        debug=config.DEBUG_MODE
    )
