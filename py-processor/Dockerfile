# Stage 1: Сборка зависимостей
FROM python:3.11-slim as builder

WORKDIR /app

# Установка системных зависимостей
RUN apt-get update && apt-get install -y \
    ffmpeg \
    tesseract-ocr \
    tesseract-ocr-rus \
    tesseract-ocr-eng \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Создание директории для кэша моделей
RUN mkdir -p /root/.cache/whisper

# Копируем requirements
COPY requirements.txt .

# Создаем и активируем виртуальное окружение (БЕЗ КАВЫЧЕК!)
RUN python -m venv /opt/venv
ENV PATH /opt/venv/bin:$PATH

# Устанавливаем зависимости
RUN pip install --no-cache-dir "numpy<2"
RUN pip install --no-cache-dir -r requirements.txt

# Генерация gRPC кода
COPY proto/ ./proto/
RUN python -m grpc_tools.protoc -I./proto --python_out=. --grpc_python_out=. ./proto/videoproc.proto

# Копируем исходный код
COPY server.py .
COPY videoproc_pb2.py .
COPY videoproc_pb2_grpc.py .

# Stage 2: Финальный образ
FROM python:3.11-slim

WORKDIR /app

# Системные зависимости (только runtime)
RUN apt-get update && apt-get install -y \
    ffmpeg \
    tesseract-ocr \
    tesseract-ocr-rus \
    tesseract-ocr-eng \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Директории для данных
RUN mkdir -p /app/tmp && \
    mkdir -p /root/.cache/whisper && \
    chmod -R 777 /app/tmp /root/.cache

# Копируем виртуальное окружение (БЕЗ КАВЫЧЕК!)
COPY --from=builder /opt/venv /opt/venv
ENV PATH /opt/venv/bin:$PATH

# Копируем приложение
COPY --from=builder /app/ ./

# Переменные окружения
ENV TESSDATA_PREFIX=/usr/share/tesseract-ocr/4.00/tessdata
ENV WHISPER_MODEL_DIR=/root/.cache/whisper
ENV GRPC_SERVER_PORT=50051

EXPOSE 50051

CMD ["python", "server.py"]