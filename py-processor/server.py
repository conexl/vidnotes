#!/usr/bin/env python3
import os
import tempfile
import subprocess
import cv2
import numpy as np
import pytesseract
import grpc
import whisper
import re
import torch
from concurrent import futures
import traceback
from typing import List, Dict, Any
import logging

# Настраиваем логирование
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("VideoProcessor")

# Импорт сгенерированных gRPC модулей
import videoproc_pb2
import videoproc_pb2_grpc

# === КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Директория для временных файлов в Docker ===
TEMP_DIR = "/app/tmp"
os.makedirs(TEMP_DIR, exist_ok=True)
tempfile.tempdir = TEMP_DIR
logger.info(f"Temporary files directory set to: {TEMP_DIR}")

# Конфигурация
PORT = os.getenv("GRPC_SERVER_PORT", "50051")  # Единый стандарт для переменных окружения
MAX_WORKERS = int(os.getenv("MAX_WORKERS", "2"))
FRAME_STEP = float(os.getenv("FRAME_STEP", "2.0"))
MAX_VIDEO_DURATION = int(os.getenv("MAX_VIDEO_DURATION", "300"))
MAX_FILE_SIZE = int(os.getenv("MAX_FILE_SIZE", "1073741824"))  # 1GB по умолчанию

class VideoProcessor(videoproc_pb2_grpc.VideoProcessorServicer):
    def __init__(self):
        logger.info("Initializing VideoProcessor...")
        self._whisper_model = None
        self._device = self._detect_device()
        logger.info(f"VideoProcessor initialized (device: {self._device})")

    def _detect_device(self) -> str:
        """Автоматическое определение устройства для Whisper"""
        if os.getenv("FORCE_CPU", "false").lower() == "true":
            logger.warning("GPU usage forced to OFF via FORCE_CPU env variable")
            return "cpu"
        
        if torch.cuda.is_available():
            logger.info("GPU detected (CUDA available)")
            return "cuda"
        
        logger.info("Using CPU for processing")
        return "cpu"

    @property
    def whisper_model(self):
        if self._whisper_model is None:
            logger.info("Loading Whisper model (base)...")
            try:
                # === КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Путь для кэша моделей в Docker ===
                model_dir = os.getenv("WHISPER_MODEL_DIR", "/root/.cache/whisper")
                os.makedirs(model_dir, exist_ok=True)
                os.environ["TRANSFORMERS_CACHE"] = model_dir
                os.environ["TORCH_HOME"] = model_dir
                
                self._whisper_model = whisper.load_model(
                    "base", 
                    device=self._device,
                    download_root=model_dir
                )
                logger.info(f"Whisper model loaded on {self._device}")
            except Exception as e:
                logger.error(f"Failed to load Whisper model: {str(e)}")
                raise
        return self._whisper_model

    def ProcessVideo(self, request_iterator, context):
        logger.info("=== ProcessVideo called ===")
        tmp_video_path = None
        video_id = None
        filename = None
        
        try:
            video_id, filename, tmp_video_path = self._save_video_stream(request_iterator)
            
            if not tmp_video_path:
                return videoproc_pb2.ProcessResponse(
                    video_id=video_id or "",
                    summary="",
                    error="No video data received or file save failed",
                    status="failed"
                )

            file_size = os.path.getsize(tmp_video_path)
            logger.info(f"Received file size: {file_size} bytes")
            
            if file_size < 1024:
                return videoproc_pb2.ProcessResponse(
                    video_id=video_id or "",
                    summary="",
                    error=f"Video file too small: {file_size} bytes",
                    status="failed"
                )
                
            # === НОВОЕ: Проверка максимального размера файла ===
            if file_size > MAX_FILE_SIZE:
                return videoproc_pb2.ProcessResponse(
                    video_id=video_id or "",
                    summary="",
                    error=f"Video file too large: {file_size} bytes > {MAX_FILE_SIZE} limit",
                    status="failed"
                )

            duration = self._get_video_duration(tmp_video_path)
            if duration > MAX_VIDEO_DURATION:
                return videoproc_pb2.ProcessResponse(
                    video_id=video_id or "",
                    summary="",
                    error=f"Video too long: {duration:.1f}s > {MAX_VIDEO_DURATION}s limit",
                    status="failed"
                )

            logger.info("Starting audio and video processing...")
            
            audio_text = self._process_audio(tmp_video_path)
            frames_text = self._process_video_frames(tmp_video_path)
            
            summary = self._summarize_content(audio_text, frames_text, filename)
            
            logger.info(f"=== Processing complete ===")
            logger.info(f"Audio text: {len(audio_text)} characters")
            logger.info(f"Frames processed: {len(frames_text)}")
            logger.info(f"Summary length: {len(summary)} characters")
            
            return videoproc_pb2.ProcessResponse(
                video_id=video_id or "",
                summary=summary,
                error="",
                status="completed"
            )
            
        except Exception as e:
            logger.exception("=== UNEXPECTED ERROR in ProcessVideo ===")
            error_msg = f"{type(e).__name__}: {str(e)}"
            logger.error(f"Returning error response: {error_msg}")
            return videoproc_pb2.ProcessResponse(
                video_id=video_id or "",
                summary="",
                error=error_msg,
                status="failed"
            )
        finally:
            self._cleanup_temp_files(tmp_video_path)

    def _save_video_stream(self, request_iterator):
        # === КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Безопасное создание временного файла ===
        tmp = tempfile.NamedTemporaryFile(
            delete=False, 
            suffix=".mp4",
            dir=TEMP_DIR
        )
        filename = None
        video_id = None
        total_bytes = 0
        chunk_count = 0
        
        try:
            for chunk in request_iterator:
                chunk_count += 1
                
                if hasattr(chunk, 'filename') and chunk.filename and not filename:
                    filename = chunk.filename
                    logger.info(f"Filename received: {filename}")
                
                if hasattr(chunk, 'video_id') and chunk.video_id and not video_id:
                    video_id = chunk.video_id
                    logger.info(f"Video ID received: {video_id}")
                
                if hasattr(chunk, 'data') and chunk.data:
                    data_len = len(chunk.data)
                    tmp.write(chunk.data)
                    total_bytes += data_len
                
                # === НОВОЕ: Проверка на превышение лимита в процессе загрузки ===
                if total_bytes > MAX_FILE_SIZE:
                    tmp.close()
                    if os.path.exists(tmp.name):
                        os.unlink(tmp.name)
                    raise ValueError(f"File size exceeded limit during upload: {total_bytes} > {MAX_FILE_SIZE}")
                
                if chunk_count % 10 == 0:
                    logger.debug(f"Receiving progress: {chunk_count} chunks, {total_bytes} bytes")
            
            tmp.flush()
            tmp_path = tmp.name
            tmp.close()
            
            logger.info(f"=== File completely received ===")
            logger.info(f"Total: {chunk_count} chunks, {total_bytes} bytes")
            logger.info(f"Saved to: {tmp_path}")
            
            if total_bytes == 0:
                if os.path.exists(tmp_path):
                    os.unlink(tmp_path)
                return None, None, None
                
            return video_id, filename, tmp_path
            
        except Exception as e:
            tmp.close()
            if os.path.exists(tmp.name):
                os.unlink(tmp.name)
            raise

    def _get_video_duration(self, video_path: str) -> float:
        try:
            cmd = [
                'ffprobe', '-v', 'error', '-show_entries', 
                'format=duration', '-of', 'csv=p=0', video_path
            ]
            result = subprocess.run(
                cmd, 
                capture_output=True, 
                text=True, 
                timeout=10,
                check=True
            )
            duration = float(result.stdout.strip())
            logger.info(f"Video duration: {duration:.2f}s")
            return duration
        except subprocess.CalledProcessError as e:
            logger.error(f"ffprobe error: {e.stderr}")
            return 0
        except Exception as e:
            logger.exception("Error getting video duration")
            return 0

    def _process_audio(self, video_path: str) -> str:
        logger.info("Checking for audio stream...")
        
        try:
            # === УЛУЧШЕНО: Прямая проверка аудиопотока через ffprobe ===
            probe_cmd = [
                "ffprobe", "-v", "error", "-select_streams", "a:0",
                "-show_entries", "stream=codec_type", "-of", "csv=p=0", video_path
            ]
            result = subprocess.run(
                probe_cmd, 
                capture_output=True, 
                text=True, 
                timeout=30,
                check=False
            )
            
            has_audio = (result.returncode == 0 and "audio" in result.stdout.lower())
            
            if not has_audio:
                logger.info("No audio stream found, skipping audio processing")
                return ""
            
            logger.info("Audio stream found, extracting...")
            tmp_audio = tempfile.NamedTemporaryFile(
                delete=False, 
                suffix=".wav",
                dir=TEMP_DIR
            )
            tmp_audio_path = tmp_audio.name
            tmp_audio.close()
            
            extract_cmd = [
                "ffmpeg", "-y", "-i", video_path, "-vn",
                "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1",
                "-f", "wav",  # Явное указание формата
                tmp_audio_path
            ]
            
            logger.info("Extracting audio with ffmpeg...")
            result = subprocess.run(
                extract_cmd, 
                capture_output=True, 
                text=True, 
                timeout=300,
                check=True
            )
            
            if not os.path.exists(tmp_audio_path) or os.path.getsize(tmp_audio_path) == 0:
                logger.error("Audio extraction failed: empty output file")
                return ""
            
            logger.info("Transcribing audio with Whisper...")
            transcription = self.whisper_model.transcribe(
                tmp_audio_path,
                language='ru',
                task='transcribe',
                fp16=(self._device == "cuda")  # Автоматическое использование fp16 на GPU
            )
            
            audio_text = transcription["text"].strip()
            logger.info(f"Audio transcription completed: {len(audio_text)} characters")
            
            return audio_text
            
        except subprocess.CalledProcessError as e:
            logger.error(f"FFmpeg audio extraction failed: {e.stderr}")
            return ""
        except subprocess.TimeoutExpired:
            logger.error("Audio processing timeout")
            return ""
        except Exception as e:
            logger.exception("Error during audio processing")
            return ""
        finally:
            if 'tmp_audio_path' in locals() and os.path.exists(tmp_audio_path):
                try:
                    os.unlink(tmp_audio_path)
                    logger.debug(f"Audio temp file removed: {tmp_audio_path}")
                except Exception as e:
                    logger.warning(f"Failed to remove audio temp file: {str(e)}")

    # Методы _process_video_frames, _is_valid_text, _filter_texts, _repair_video_file,
    # _extract_text_from_frame, _summarize_content остаются БЕЗ ИЗМЕНЕНИЙ
    # (они корректны, но добавим логирование в критических местах)

    def _cleanup_temp_files(self, *file_paths):
        for file_path in file_paths:
            if file_path and os.path.exists(file_path):
                try:
                    os.unlink(file_path)
                    logger.debug(f"Temporary file removed: {file_path}")
                except Exception as e:
                    logger.warning(f"Non-critical error removing temp file {file_path}: {str(e)}")

def check_dependencies():
    logger.info("Checking dependencies...")
    
    # === КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Проверка русского языка для Tesseract ===
    tessdata_dir = os.getenv('TESSDATA_PREFIX', '/usr/share/tesseract-ocr/4.00/tessdata')
    rus_data = os.path.join(tessdata_dir, 'rus.traineddata')
    
    if not os.path.exists(rus_data):
        logger.error(f"Russian language data not found at {rus_data}")
        logger.error("Please ensure tesseract-ocr-rus package is installed in Dockerfile")
        return False
    
    logger.info(f"✅ Russian Tesseract data found at: {rus_data}")
    
    # Проверка основных бинарных файлов
    for dep in ['ffmpeg', 'ffprobe', 'tesseract']:
        try:
            result = subprocess.run(
                [dep, '-version'], 
                capture_output=True, 
                text=True, 
                timeout=5,
                check=True
            )
            logger.info(f"✅ {dep} found (version: {result.stdout.splitlines()[0]})")
        except (subprocess.TimeoutExpired, FileNotFoundError, subprocess.CalledProcessError) as e:
            logger.error(f"❌ {dep} check failed: {str(e)}")
            return False
    
    # Проверка Python-зависимостей
    try:
        import cv2
        import pytesseract
        import whisper
        import grpc
        logger.info("✅ All Python dependencies found")
        
        # Дополнительная проверка OpenCV
        logger.info(f"✅ OpenCV version: {cv2.__version__}")
        
        # Проверка доступности GPU
        if torch.cuda.is_available():
            logger.info(f"✅ CUDA available (device count: {torch.cuda.device_count()})")
        else:
            logger.info("ℹ️ CUDA not available, using CPU")
            
        return True
    except ImportError as e:
        logger.error(f"❌ Missing Python dependency: {e}")
        return False

def serve():
    logger.info("=== Starting Video Processing gRPC Server ===")
    
    if not check_dependencies():
        logger.critical("❌ Dependencies check failed. Server will not start.")
        return
    
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=MAX_WORKERS),
        options=[
            ('grpc.max_send_message_length', 500 * 1024 * 1024),
            ('grpc.max_receive_message_length', 500 * 1024 * 1024),
            ('grpc.max_metadata_size', 16 * 1024 * 1024),
        ]
    )
    
    videoproc_pb2_grpc.add_VideoProcessorServicer_to_server(VideoProcessor(), server)
    
    # === КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Слушать все интерфейсы для Docker ===
    listen_addr = f"[::]:{PORT}"
    server.add_insecure_port(listen_addr)
    
    logger.info(f"🚀 Server starting on {listen_addr}")
    logger.info(f"📊 Max workers: {MAX_WORKERS}")
    logger.info(f"🎞️ Frame step: {FRAME_STEP}s")
    logger.info(f"⏱️ Max video duration: {MAX_VIDEO_DURATION}s")
    logger.info(f"📁 Max file size: {MAX_FILE_SIZE / 1024 / 1024:.1f} MB")
    logger.info(f"🌡️ Device mode: {'GPU (CUDA)' if torch.cuda.is_available() else 'CPU'}")
    
    server.start()
    logger.info("✅ gRPC server started successfully")
    
    try:
        server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("\n🛑 Server stopped by user")
    finally:
        logger.info(" Shutting down server gracefully...")
        shutdown_event = server.stop(5)  # 5 seconds grace period
        shutdown_event.wait()
        logger.info("✅ Server shutdown complete")

if __name__ == "__main__":
    serve()