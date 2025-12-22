FROM --platform=linux/amd64 nvidia/cuda:13.0.2-cudnn-devel-ubuntu24.04

ENV DEBIAN_FRONTEND=noninteractive \
    PYTHONUNBUFFERED=1 \
    CUDA_HOME=/usr/local/cuda \
    PATH=/usr/local/cuda/bin:${PATH} \
    LD_LIBRARY_PATH=/usr/local/cuda/lib64:/usr/local/cuda/extras/CUPTI/lib64:/usr/local/cuda/targets/x86_64-linux/lib:${LD_LIBRARY_PATH} \
    PIP_NO_CACHE_DIR=1

RUN rm -f /etc/apt/sources.list.d/cuda*.list /etc/apt/sources.list.d/nvidia-*.list

RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        git \
        curl \
        ffmpeg \
        libgl1 \
        libglib2.0-0 \
        libsm6 \
        libxrender1 \
        libxext6 \
        libjpeg-dev \
        libpng-dev \
        zlib1g-dev \
        python3.12 \
        python3.12-dev \
        python3.12-venv \
        python3-pip \
    && update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.12 1 \
    && update-alternatives --install /usr/bin/python python /usr/bin/python3.12 1 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV VIRTUAL_ENV=/opt/venv
RUN python3.12 -m venv $VIRTUAL_ENV
ENV PATH="$VIRTUAL_ENV/bin:$PATH"

ENV LD_LIBRARY_PATH=$VIRTUAL_ENV/lib/python3.12/site-packages/nvidia/cublas/lib:${LD_LIBRARY_PATH}

RUN pip install --upgrade pip setuptools wheel

RUN pip install --no-cache-dir \
    torch==2.5.1+cu121 \
    torchvision==0.20.1+cu121 \
    torchaudio==2.5.1+cu121 \
    --index-url https://download.pytorch.org/whl/cu121

RUN set -eux; \
    apt-get update && apt-get install -y --no-install-recommends gnupg curl ca-certificates || true; \
    if curl -fsSL https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/x86_64/3bf863cc.pub -o /tmp/nvidia_cuda_pub.pem; then \
        apt-key add /tmp/nvidia_cuda_pub.pem || true; \
    else \
    fi; \
    if curl -fsSL https://developer.download.nvidia.com/compute/machine-learning/repos/ubuntu2004/x86_64/nvidia-machine-learning-repo.pub -o /tmp/nvidia_ml_pub.pem; then \
        apt-key add /tmp/nvidia_ml_pub.pem || true; \
    else \
    fi; \
    apt-get update || true; \
    if ! apt-get install -y --no-install-recommends libnvinfer8 libnvinfer-dev libnvinfer-plugin8 libnvinfer-bin; then \
    else \
        apt-get clean; rm -rf /var/lib/apt/lists/*; ldconfig; \
    fi; \
    rm -f /tmp/nvidia_cuda_pub.pem /tmp/nvidia_ml_pub.pem || true

RUN pip install --no-cache-dir \
    tensorrt==10.7.0 \
    tensorrt-cu12==10.7.0 \
    tensorrt-cu12-bindings==10.7.0 \
    tensorrt-cu12-libs==10.7.0 \
    --extra-index-url https://pypi.nvidia.com || true

ENV LD_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu:${LD_LIBRARY_PATH}

RUN pip install --no-cache-dir onnxruntime-gpu==1.23.2

RUN pip install --no-cache-dir tensorflow==2.20.0

WORKDIR /app

COPY requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir -r /app/requirements.txt

COPY . /app

EXPOSE 5100

HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://127.0.0.1:5100/health || exit 1

CMD ["bash", "/app/entrypoint.sh"]