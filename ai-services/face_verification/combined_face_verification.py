import cv2
import numpy as np
from deepface import DeepFace
from insightface.app import FaceAnalysis
try:
    from insightface.model_zoo import get_model
    HAS_LIVENESS = True
except:
    HAS_LIVENESS = False
    print("Warning: Liveness model not available, skipping liveness detection")
import os
import tempfile
from collections import deque
from concurrent.futures import ThreadPoolExecutor, as_completed
import threading

class CombinedFaceVerification:
    def __init__(self, 
                 base_image_path,
                 insightface_similarity_threshold=None,
                 deepface_similarity_threshold=None,
                 spoof_threshold=None,
                 frame_interval=None,
                 insightface_app=None,
                 liveness_model=None):
        
        import config

        self.base_image_path = base_image_path

        self.insightface_sim_threshold = insightface_similarity_threshold if insightface_similarity_threshold is not None else config.INSIGHTFACE_SIM_THRESHOLD
        self.deepface_sim_threshold = deepface_similarity_threshold if deepface_similarity_threshold is not None else config.DEEPFACE_SIM_THRESHOLD
        self.spoof_threshold = spoof_threshold if spoof_threshold is not None else config.SPOOF_RATE_THRESHOLD
        self.frame_interval = frame_interval if frame_interval is not None else config.FRAME_INTERVAL
        
        self.max_workers = getattr(config, 'MAX_WORKERS', min(8, (os.cpu_count() or 1)))
        
        self._temp_file_lock = threading.Lock()
        self._temp_files = []
        
        import config
        import logging
        logger = logging.getLogger(__name__)

        if insightface_app is not None:
            self.insightface_app = insightface_app
            self.liveness_model = liveness_model
            # infer whether the provided app was prepared on GPU; fall back to config
            use_gpu = getattr(insightface_app, '_prepared_on_gpu', None)
            if use_gpu is None:
                use_gpu = config.USE_GPU
            logger.info(f"ℹ️  Using preloaded InsightFace and liveness model instances (prepared_on_gpu={use_gpu})")
        else:
            use_gpu = config.USE_GPU
            if use_gpu:
                try:
                    import onnxruntime
                    available_providers = onnxruntime.get_available_providers()
                    if 'CUDAExecutionProvider' not in available_providers and 'TensorrtExecutionProvider' not in available_providers:
                        logger.info("⚠️  GPU requested but CUDA not available, falling back to CPU")
                        use_gpu = False
                    else:
                        logger.info(f"✅ GPU Mode ENABLED - Available providers: {available_providers}")
                except Exception as e:
                    logger.info(f"⚠️  Could not check GPU availability: {e}, falling back to CPU")
                    use_gpu = False
            else:
                logger.info("ℹ️  CPU Mode - GPU is disabled")

            if use_gpu:
                try:
                    import onnxruntime
                    available_providers = onnxruntime.get_available_providers()
                except Exception:
                    available_providers = []

                providers = []
                if 'CUDAExecutionProvider' in available_providers:
                    providers.append(('CUDAExecutionProvider', {
                        'device_id': config.GPU_ID,
                        'gpu_mem_limit': getattr(config, 'CUDA_GPU_MEM_LIMIT', 16 * 1024 * 1024 * 1024),
                        'arena_extend_strategy': getattr(config, 'CUDA_ARENA_EXTEND_STRATEGY', 'kSameAsRequested'),
                    }))

                if config.USE_TENSORRT and 'TensorrtExecutionProvider' in available_providers:
                    providers.append(('TensorrtExecutionProvider', {
                        'device_id': config.GPU_ID,
                        'trt_fp16_enable': config.USE_FP16,
                        'trt_max_workspace_size': config.MAX_WORKSPACE_SIZE,
                    }))

                if not providers:
                    providers = ['CPUExecutionProvider']
            else:
                providers = ['CPUExecutionProvider']

            self.insightface_app = FaceAnalysis(providers=providers)
            self.insightface_app.prepare(ctx_id=config.GPU_ID if use_gpu else -1, det_size=config.INSIGHTFACE_DET_SIZE)

            self.liveness_model = None
            if HAS_LIVENESS:
                try:
                    self.liveness_model = get_model(config.LIVENESS_MODEL_NAME)
                    if self.liveness_model:
                        ctx_id = config.GPU_ID if use_gpu else -1
                        self.liveness_model.prepare(ctx_id=ctx_id, input_size=(128, 128))
                        logger.info(f"✅ Liveness model loaded on {'GPU' if use_gpu else 'CPU'}")
                except Exception as e:
                    logger.warning(f"⚠️  Could not load liveness model: {e}")
        
        # Ensure CPU-backed models always use batch size 1 for stability
        self.batch_size = 1 if not use_gpu else config.BATCH_SIZE
        logger.info(f"📊 Face verification initialized - Mode: {'GPU (CUDA)' if use_gpu else 'CPU'} | Batch: {self.batch_size} | FP16: {config.USE_FP16}")
        
        self.MOTION_BUFFER = 5
        self.MOTION_DIFF_THRESHOLD = 2.0
        # Liveness score threshold from config
        self.LIVENESS_THRESHOLD = getattr(__import__('config'), 'LIVENESS_SCORE_THRESHOLD', 0.9)
        self.motion_buffers = {}
        
        base_img = cv2.imread(base_image_path)
        base_faces = self.insightface_app.get(base_img)
        if not base_faces:
            raise ValueError("No face found in base image (InsightFace)")
        if len(base_faces) > 1:
            raise ValueError(f"Expected exactly one face in base image, found {len(base_faces)}")
        self.insightface_base_embedding = base_faces[0].embedding
        

        base_faces_deepface = DeepFace.extract_faces(img_path=base_image_path, enforce_detection=True)
        if len(base_faces_deepface) == 0:
            raise ValueError("No face detected in base image (DeepFace)")
        if len(base_faces_deepface) > 1:
            raise ValueError(f"❌ ERROR: Multiple faces detected in base image ({len(base_faces_deepface)} faces)! Base image must contain exactly ONE face.")
    
    def _get_temp_file(self):
        with self._temp_file_lock:
            temp_file = tempfile.NamedTemporaryFile(suffix='.jpg', delete=False)
            temp_path = temp_file.name
            temp_file.close()
            self._temp_files.append(temp_path)
            return temp_path
    
    def _cleanup_temp_files(self):
        with self._temp_file_lock:
            for path in self._temp_files:
                try:
                    if os.path.exists(path):
                        os.unlink(path)
                except:
                    pass
            self._temp_files.clear()
    
    def get_insightface_results_batch(self, frames, frame_numbers):
        import logging
        logger = logging.getLogger(__name__)
        
        batch_results = []
        
        all_faces = []
        for frame in frames:
            faces = self.insightface_app.get(frame)
            all_faces.append(faces)
        
        face_crops_for_liveness = []
        face_indices = [] 
        
        for i, (faces, frame, frame_num) in enumerate(zip(all_faces, frames, frame_numbers)):
            multiple_faces = False
            
            if not faces:
                batch_results.append((None, None, None, multiple_faces))
                continue
            
            if len(faces) > 1:
                multiple_faces = True
                logger.warning(f"⚠️  Multiple faces detected in frame {frame_num} (InsightFace)")
                batch_results.append((None, None, None, multiple_faces))
                continue
            
            face = faces[0]
            bbox = face.bbox.astype(int)
            x1, y1, x2, y2 = bbox
            face_crop = frame[y1:y2, x1:x2]
            
            cos_sim = np.dot(self.insightface_base_embedding, face.embedding) / (
                np.linalg.norm(self.insightface_base_embedding) * np.linalg.norm(face.embedding)
            )
            
            if self.liveness_model:
                h, w = face_crop.shape[:2]
                if h > 0 and w > 0:
                    face_crops_for_liveness.append((face_crop, bbox, frame_num))
                    face_indices.append(i)
            
            batch_results.append((cos_sim, None, None, False, face_crop, bbox, frame_num))
        
        if self.liveness_model and face_crops_for_liveness:
            try:
                resized_batch = []
                for face_crop, _, _ in face_crops_for_liveness:
                    resized = cv2.resize(face_crop, (128, 128))
                    resized_rgb = cv2.cvtColor(resized, cv2.COLOR_BGR2RGB)
                    resized_batch.append(resized_rgb)
                
                liveness_scores = []
                for resized_rgb in resized_batch:
                    result = self.liveness_model.get(resized_rgb)
                    liveness_scores.append(float(result) if result is not None else None)
                
                for idx, (face_idx, liveness_score) in enumerate(zip(face_indices, liveness_scores)):
                    if liveness_score is not None:
                        is_real = liveness_score > self.LIVENESS_THRESHOLD
                        cos_sim, _, _, mult, face_crop, bbox, frame_num = batch_results[face_idx]
                        batch_results[face_idx] = (cos_sim, liveness_score, is_real, mult)
                    
            except Exception as e:
                logger.debug(f"Batch liveness model error: {e}")
        
        final_results = []
        for i, result in enumerate(batch_results):
            if len(result) == 4:
                final_results.append(result)
            else:
                cos_sim, liveness_score, is_real, mult, face_crop, bbox, frame_num = result
                
                if is_real is None:
                    # Apply motion-based liveness
                    try:
                        x1, y1, x2, y2 = bbox
                        cx = int((x1 + x2) / 2) // 20
                        cy = int((y1 + y2) / 2) // 20
                        face_key = (cx, cy)
                        
                        gray = cv2.cvtColor(face_crop, cv2.COLOR_BGR2GRAY)
                        h, w = gray.shape
                        if h > 0 and w > 0:
                            small = cv2.resize(gray, (64, 64))
                            
                            buf = self.motion_buffers.get(face_key)
                            if buf is None:
                                self.motion_buffers[face_key] = {
                                    'last_crop': small, 
                                    'diffs': deque(maxlen=self.MOTION_BUFFER)
                                }
                                liveness_score = 0.0
                                is_real = False
                            else:
                                last = buf['last_crop']
                                diff = np.mean(np.abs(small.astype(np.float32) - last.astype(np.float32)))
                                buf['diffs'].append(diff)
                                buf['last_crop'] = small
                                avg_diff = float(np.mean(buf['diffs'])) if len(buf['diffs']) > 0 else 0.0
                                liveness_score = avg_diff
                                is_real = avg_diff > self.MOTION_DIFF_THRESHOLD
                    except Exception as e:
                        logger.debug(f"Motion-based liveness error: {e}")
                
                final_results.append((cos_sim, liveness_score, is_real, mult))
        
        return final_results
    
    def get_deepface_spoof_results(self, frame, frame_count):
        import logging
        logger = logging.getLogger(__name__)
        
        temp_path = self._get_temp_file()
        
        try:
            success = cv2.imwrite(temp_path, frame)
            if not success:
                return None
            
            if not os.path.exists(temp_path) or os.path.getsize(temp_path) == 0:
                return None
            
            frame_faces = DeepFace.extract_faces(
                img_path=temp_path, 
                anti_spoofing=True, 
                enforce_detection=False
            )
            
            if len(frame_faces) == 0:
                return None
            
            if len(frame_faces) > 1:
                logger.warning(f"Multiple faces detected in frame {frame_count} ({len(frame_faces)} faces)")
                # return False to indicate multiple faces detected (legacy behavior)
                return False  
            
            is_real = all(face["is_real"] for face in frame_faces)
            return is_real
            
        except Exception as e:
            logger.debug(f"DeepFace extraction error: {e}")
            return None
    
    def combine_spoof_scores(self, deepface_is_real):
        if deepface_is_real is None:
            return None
        return deepface_is_real
    
    def combine_verification_scores(self, insightface_sim):
        if insightface_sim is None:
            return None
        return insightface_sim
    
    def _process_frame_batch(self, frame_batch, frame_numbers, results, messages):
        batch_processed = 0
        batch_real = 0
        batch_verified = 0
        batch_spoofed = 0
        batch_two_faces = 0

        insightface_results = self.get_insightface_results_batch(frame_batch, frame_numbers)
        
        deepface_results_dict = {}
        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            deepface_futures = {
                executor.submit(self.get_deepface_spoof_results, frame, frame_num): frame_num
                for frame, frame_num in zip(frame_batch, frame_numbers)
            }
            
            for future in as_completed(deepface_futures):
                frame_num = deepface_futures[future]
                try:
                    result = future.result()
                    deepface_results_dict[frame_num] = result
                except Exception as e:
                    import logging
                    logger = logging.getLogger(__name__)
                    logger.debug(f"DeepFace thread error for frame {frame_num}: {e}")
                    deepface_results_dict[frame_num] = None
        
        for i, frame_num in enumerate(frame_numbers):
            insightface_sim, insightface_liveness, insightface_real, insightface_multiple = insightface_results[i]
            deepface_is_real = deepface_results_dict.get(frame_num)
            
            multiple_faces_detected = insightface_multiple or (deepface_is_real is False)
            
            if multiple_faces_detected:
                batch_processed += 1
                batch_two_faces += 1
                messages.append({'frame': frame_num, 'error': 'multiple_faces_detected'})
                continue

            if insightface_sim is None and deepface_is_real is None:
                continue

            batch_processed += 1

            is_real = deepface_is_real if deepface_is_real is not None else insightface_real
            is_verified = insightface_sim is not None and insightface_sim >= self.insightface_sim_threshold

            if is_real:
                batch_real += 1
                if is_verified:
                    batch_verified += 1
            else:
                batch_spoofed += 1

            results.append({
                'frame': frame_num,
                'is_real': bool(is_real) if is_real is not None else None,
                'is_verified': bool(is_verified),
                'insightface_sim': float(insightface_sim) if insightface_sim is not None else None,
                'insightface_liveness': float(insightface_liveness) if insightface_liveness is not None else None
            })
        
        return {
            'processed': batch_processed,
            'real': batch_real,
            'verified': batch_verified,
            'spoofed': batch_spoofed,
            'two_faces': batch_two_faces
        }    
    

    def process_video(self, video_path):
        import config
        
        if not os.path.exists(video_path):
            return {
                'success': False,
                'reason': 'video_not_found',
                'message': f'Video file not found: {video_path}',
                'stats': None,
                'results': []
            }

        cap = cv2.VideoCapture(video_path)
        if not cap.isOpened():
            return {
                'success': False,
                'reason': 'cannot_open_video',
                'message': 'Cannot open video file',
                'stats': None,
                'results': []
            }

        total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
        fps = cap.get(cv2.CAP_PROP_FPS)
        width = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
        height = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
        duration = total_frames / fps if fps > 0 else 0

        frame_count = 0
        processed_frames = 0
        real_frames = 0
        verified_frames = 0
        spoofed_frames = 0
        two_faces_frames = 0
        results = []
        messages = []

        batch_size = self.batch_size
        frame_batch = []
        frame_numbers = []

        try:
            while True:
                ret, frame = cap.read()
                if not ret:
                    if len(frame_batch) > 0:
                        stats = self._process_frame_batch(frame_batch, frame_numbers, results, messages)
                        processed_frames += stats['processed']  
                        real_frames += stats['real']            
                        verified_frames += stats['verified']    
                        spoofed_frames += stats['spoofed']      
                        two_faces_frames += stats['two_faces']  
                    break

                frame_count += 1

                if frame_count % self.frame_interval != 0:
                    continue

                frame_batch.append(frame)
                frame_numbers.append(frame_count)

                if len(frame_batch) >= batch_size:
                    stats = self._process_frame_batch(frame_batch, frame_numbers, results, messages)
                    processed_frames += stats['processed']  
                    real_frames += stats['real']            
                    verified_frames += stats['verified']    
                    spoofed_frames += stats['spoofed']      
                    two_faces_frames += stats['two_faces']  
                    frame_batch = []
                    frame_numbers = []

        finally:
            cap.release()
            self._cleanup_temp_files()

        if processed_frames == 0:
            return {
                'success': False,
                'reason': 'no_faces_processed',
                'message': 'No frames with faces were processed',
                'stats': {
                    'total_frames': frame_count,
                    'processed_frames': processed_frames,
                    'real_frames': real_frames,
                    'spoofed_frames': spoofed_frames,
                    'verified_frames': verified_frames
                },
                'results': results,
                'messages': messages
            }

        spoof_rate = (spoofed_frames / processed_frames) * 100
        real_rate = (real_frames / processed_frames) * 100
        verification_rate = (verified_frames / real_frames) * 100 if real_frames > 0 else 0
        two_faces_rate = (two_faces_frames / processed_frames) * 100 if processed_frames > 0 else 0

        import logging
        logger = logging.getLogger(__name__)
        
        try:
            import config
            spoof_threshold_pct = config.SPOOF_RATE_THRESHOLD * 100
            verification_threshold_pct = config.VERIFICATION_THRESHOLD * 100
            two_faces_threshold_pct = getattr(config, 'TWO_FACES_FRAME_THRESHOLD', 0.40) * 100
        except Exception as e:
            logger.error(f"❌ Error loading config: {e}")
            spoof_threshold_pct = 30
            verification_threshold_pct = 70
            two_faces_threshold_pct = 40
        
        logger.info(f"🔍 Thresholds - Spoof: {spoof_threshold_pct:.1f}%, Verification: {verification_threshold_pct:.1f}%, Two Faces: {two_faces_threshold_pct:.1f}%")
        logger.info(f"📊 Results - Spoof Rate: {spoof_rate:.1f}%, Verification Rate: {verification_rate:.1f}%, Two Faces Rate: {two_faces_rate:.1f}%")

        stats = {
            'total_frames': frame_count,
            'fps': float(fps) if fps is not None else None,
            'width': width,
            'height': height,
            'duration': float(duration),
            'processed_frames': processed_frames,
            'real_frames': real_frames,
            'spoofed_frames': spoofed_frames,
            'two_faces_frames': two_faces_frames,
            'verified_frames': verified_frames,
            'spoof_rate': float(spoof_rate),
            'real_rate': float(real_rate),
            'verification_rate': float(verification_rate),
            'two_faces_rate': float(two_faces_rate)
        }

        # Decision logic
        if spoof_rate > spoof_threshold_pct:
            return {
                'success': False,
                'reason': 'high_spoof_rate',
                'message': 'High spoofing detected',
                'stats': stats,
                'results': results,
                'messages': messages
            }
        if two_faces_rate > two_faces_threshold_pct:
            return {
                'success': False,
                'reason': 'multiple_faces_exceeded',
                'message': f'Video contains too many frames with multiple faces ({two_faces_rate:.1f}% > {two_faces_threshold_pct:.1f}%)',
                'stats': stats,
                'results': results,
                'messages': messages
            }
        if real_frames == 0:
            return {
                'success': False,
                'reason': 'no_real_faces',
                'message': 'No real faces found',
                'stats': stats,
                'results': results,
                'messages': messages
            }
        if verification_rate >= verification_threshold_pct:
            return {
                'success': True,
                'reason': 'verified',
                'message': 'Verification successful',
                'stats': stats,
                'results': results,
                'messages': messages
            }

        return {
            'success': False,
            'reason': 'insufficient_match',
            'message': f'Insufficient match: {verification_rate:.1f}% < 70% ',
            'stats': stats,
            'results': results,
            'messages': messages
        }
