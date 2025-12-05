#!/bin/bash
set -e

echo "Starting Face Verification Service..."
echo "Python version: $(python3 --version)"
echo "PyTorch version: $(python3 -c 'import torch; print(torch.__version__)')"
echo "CUDA available: $(python3 -c 'import torch; print(torch.cuda.is_available())')"
echo "GPU Enabled: ${USE_GPU:-true}"
echo "GPU ID: ${GPU_ID:-0}"
echo "Port: ${PORT:-5100}"

exec python3 server.py
