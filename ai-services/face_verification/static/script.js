const API_BASE = 'http://localhost:5100';
let croppedImageBlob = null;

document.addEventListener('DOMContentLoaded', () => {
    checkHealth();
    loadConfig();
    setupEventListeners();
});

async function checkHealth() {
    try {
        const res = await fetch(`${API_BASE}/health`);
        const data = await res.json();
        
        const statusEl = document.getElementById('health-status');
        const versionEl = document.getElementById('service-version');
        
        statusEl.textContent = '🟢 OK';
        statusEl.classList.add('ok');
        versionEl.textContent = data.version;
    } catch (e) {
        const statusEl = document.getElementById('health-status');
        statusEl.textContent = '🔴 Offline';
        statusEl.classList.add('error');
        console.error('Health check failed:', e);
    }
}

async function loadConfig() {
    try {
        // Fetch info (limits & thresholds) and current config
        const [infoRes, cfgRes] = await Promise.all([
            fetch(`${API_BASE}/info`),
            fetch(`${API_BASE}/config`)
        ]);

        const info = await infoRes.json();
        const cfg = await cfgRes.json();

        const configContent = document.getElementById('config-content');

        let html = '<table class="config-table">';

        html += '<tr><td colspan="3" style="background: #f5f5f5; font-weight: bold;">Limits</td></tr>';
        html += `<tr><td>Max Image Size</td><td colspan="2">${info.limits.max_image_size_mb}MB</td></tr>`;
        html += `<tr><td>Max Video Size</td><td colspan="2">${info.limits.max_video_size_mb}MB</td></tr>`;
        html += `<tr><td>Image Types</td><td colspan="2">${info.limits.allowed_image_types.join(', ')}</td></tr>`;
        html += `<tr><td>Video Types</td><td colspan="2">${info.limits.allowed_video_types.join(', ')}</td></tr>`;

        html += '<tr><td colspan="3" style="background: #f5f5f5; font-weight: bold;">Runtime Configuration (editable)</td></tr>';

        const params = {
            'frame_interval': { value: cfg.processing.frame_interval, type: 'number' },
            'insightface_similarity_threshold': { value: cfg.processing.insightface_similarity_threshold, type: 'number', step: '0.01' },
            'deepface_similarity_threshold': { value: cfg.processing.deepface_similarity_threshold, type: 'number', step: '0.01' },
            'verification_threshold': { value: cfg.processing.verification_threshold, type: 'number', step: '0.01' },
            'spoof_rate_threshold': { value: cfg.processing.spoof_rate_threshold, type: 'number', step: '0.01' },
            'use_gpu': { value: cfg.hardware.use_gpu, type: 'boolean' },
            'gpu_id': { value: cfg.hardware.gpu_id, type: 'number' },
            'use_fp16': { value: cfg.hardware.use_fp16, type: 'boolean' },
            'batch_size': { value: cfg.hardware.batch_size, type: 'number' },
            'use_tensorrt': { value: cfg.hardware.use_tensorrt, type: 'boolean' },
            'liveness_score_threshold': { value: cfg.liveness.score_threshold, type: 'number', step: '0.01' },
            'cuda_gpu_mem_limit': { value: cfg.hardware.cuda_gpu_mem_limit, type: 'number' },
            'cuda_arena_extend_strategy': { value: cfg.hardware.cuda_arena_extend_strategy || 'kSameAsRequested', type: 'string' }
        };

        for (const [key, meta] of Object.entries(params)) {
            const displayKey = key.replace(/_/g, ' ');
            let inputHtml = '';
            if (meta.type === 'boolean') {
                const checked = meta.value ? 'checked' : '';
                inputHtml = `<input type="checkbox" id="cfg-${key}" ${checked}>`;
            } else {
                const step = meta.step ? `step="${meta.step}"` : '';
                if (meta.type === 'string') {
                    inputHtml = `<input type="text" id="cfg-${key}" value="${meta.value}">`;
                } else {
                    inputHtml = `<input type="number" id="cfg-${key}" value="${meta.value}" ${step}>`;
                }
            }

            html += `<tr><td>${displayKey}</td><td>${inputHtml}</td><td><button class="btn btn-secondary" onclick="updateConfigParam('${key}')">Update</button></td></tr>`;
        }

        html += '</table>';
        configContent.innerHTML = html;
    } catch (e) {
        console.error('Config load failed:', e);
        document.getElementById('config-content').innerHTML = '<p style="color: red;">Failed to load config</p>';
    }
}

// Update a single config parameter using PATCH /config/<param>
async function updateConfigParam(param) {
    const el = document.getElementById(`cfg-${param}`);
    if (!el) return;

    let value;
    if (el.type === 'checkbox') {
        value = el.checked;
    } else {
        // number input
        const raw = el.value;
        // try to parse number; if empty show error
        if (raw === '' || raw === null) {
            alert('Value cannot be empty');
            return;
        }
        // If the input contains a dot or step defined, treat as float for thresholds
        value = raw.includes('.') ? parseFloat(raw) : parseInt(raw, 10);
        if (Number.isNaN(value)) {
            alert('Invalid numeric value');
            return;
        }
    }

    const btn = event?.target || null;
    if (btn) btn.disabled = true;

    try {
        const res = await fetch(`${API_BASE}/config/${param}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ value })
        });

        const data = await res.json();
        if (res.ok) {
            // reflect updated value (reload config)
            await loadConfig();
            alert(`${param} updated successfully`);
        } else {
            alert(`Failed to update: ${data.error || data.message}`);
        }
    } catch (e) {
        console.error('Config update error:', e);
        alert('Failed to update configuration (connection error)');
    } finally {
        if (btn) btn.disabled = false;
    }
}

function setupEventListeners() {
    document.querySelectorAll('.tab-button').forEach(btn => {
        btn.addEventListener('click', (e) => {
            switchTab(e.target.dataset.tab);
        });
    });

    // Crop tab
    document.getElementById('crop-input').addEventListener('change', previewCropImage);
    document.getElementById('crop-btn').addEventListener('click', cropImage);
    document.getElementById('crop-download-btn')?.addEventListener('click', downloadCropped);

    // Verify tab
    document.getElementById('verify-photo').addEventListener('change', updatePhotoName);
    document.getElementById('verify-video').addEventListener('change', updateVideoName);
    document.getElementById('verify-btn').addEventListener('click', verifyFace);
}


function switchTab(tabName) {
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });

    document.querySelectorAll('.tab-button').forEach(btn => {
        btn.classList.remove('active');
    });

    document.getElementById(tabName).classList.add('active');
    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
}


function previewCropImage(e) {
    const file = e.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
        const preview = document.getElementById('crop-preview');
        preview.innerHTML = `<img src="${event.target.result}" alt="preview">`;
    };
    reader.readAsDataURL(file);
}

async function cropImage() {
    const input = document.getElementById('crop-input');
    if (!input.files[0]) {
        alert('Please select an image');
        return;
    }

    const btn = document.getElementById('crop-btn');
    const resultDiv = document.getElementById('crop-result');
    
    btn.disabled = true;
    btn.innerHTML = '<span class="loading"></span> Processing...';
    resultDiv.classList.add('hidden');

    const formData = new FormData();
    formData.append('image', input.files[0]);

    try {
        const res = await fetch(`${API_BASE}/crop`, {
            method: 'POST',
            body: formData
        });

        if (res.ok) {
            // Success - image returned
            croppedImageBlob = await res.blob();
            
            const reader = new FileReader();
            reader.onload = (event) => {
                const preview = document.getElementById('crop-preview');
                preview.innerHTML = `<img src="${event.target.result}" alt="cropped">`;
                
                document.getElementById('crop-title').textContent = '✅ Crop Successful!';
                document.getElementById('crop-message').textContent = 'Image has been cropped to passport style.';
                document.getElementById('crop-details').innerHTML = '';
                
                resultDiv.classList.remove('hidden', 'error');
                resultDiv.classList.add('success');
            };
            reader.readAsDataURL(croppedImageBlob);
        } else {
            const error = await res.json();
            showCropError(error);
        }
    } catch (e) {
        console.error('Crop error:', e);
        showCropError({ error: 'Connection error', message: e.message });
    } finally {
        btn.disabled = false;
        btn.textContent = 'Crop Image';
    }
}

function showCropError(error) {
    const resultDiv = document.getElementById('crop-result');
    document.getElementById('crop-title').textContent = `❌ Error: ${error.error}`;
    document.getElementById('crop-message').textContent = error.message;
    document.getElementById('crop-details').innerHTML = '';
    
    resultDiv.classList.remove('hidden', 'success');
    resultDiv.classList.add('error');
}

function downloadCropped() {
    if (!croppedImageBlob) return;
    
    const url = URL.createObjectURL(croppedImageBlob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'cropped.jpg';
    a.click();
    URL.revokeObjectURL(url);
}


function updatePhotoName() {
    const file = document.getElementById('verify-photo').files[0];
    document.getElementById('photo-file-name').textContent = file ? file.name : 'No file selected';
}

function updateVideoName() {
    const file = document.getElementById('verify-video').files[0];
    document.getElementById('video-file-name').textContent = file ? file.name : 'No file selected';
}

async function verifyFace() {
    const photoInput = document.getElementById('verify-photo');
    const videoInput = document.getElementById('verify-video');

    if (!photoInput.files[0] || !videoInput.files[0]) {
        alert('Please select both photo and video');
        return;
    }

    const btn = document.getElementById('verify-btn');
    const resultDiv = document.getElementById('verify-result');
    
    btn.disabled = true;
    btn.innerHTML = '<span class="loading"></span> Verifying... (this may take a minute)';
    resultDiv.classList.add('hidden');

    const formData = new FormData();
    formData.append('photo', photoInput.files[0]);
    formData.append('video', videoInput.files[0]);

    try {
        const res = await fetch(`${API_BASE}/verify`, {
            method: 'POST',
            body: formData
        });

        const data = await res.json();

        if (data.success) {
            showVerifySuccess(data);
        } else {
            showVerifyError(data);
        }

        resultDiv.classList.remove('hidden');
    } catch (e) {
        console.error('Verify error:', e);
        showVerifyError({ 
            reason: 'connection_error',
            message: `Connection failed: ${e.message}`
        });
        resultDiv.classList.remove('hidden');
    } finally {
        btn.disabled = false;
        btn.textContent = 'Verify';
    }
}

function showVerifySuccess(data) {
    const resultDiv = document.getElementById('verify-result');
    
    document.getElementById('verify-title').textContent = '✅ ' + data.message;
    document.getElementById('verify-message').textContent = `Reason: ${data.reason}`;
    
    // Stats
    const stats = data.stats;
    if (stats) {
        document.getElementById('stat-verification').textContent = stats.verification_rate.toFixed(1) + '%';
        document.getElementById('stat-spoof').textContent = stats.spoof_rate.toFixed(1) + '%';
        document.getElementById('stat-real').textContent = stats.real_rate.toFixed(1) + '%';
        document.getElementById('stat-frames').textContent = stats.processed_frames;
        document.getElementById('verify-stats').classList.remove('hidden');
    }
    
    const details = formatJson(stats);
    document.getElementById('verify-details').innerHTML = `<pre>${details}</pre>`;
    
    // Frame results
    const frameHtml = data.results.map(f => {
        const cssClass = f.is_verified ? 'verified' : f.is_real ? 'real' : 'spoofed';
        const status = f.is_verified ? '✓ Verified' : f.is_real ? '~ Real' : '✗ Spoofed';
        const sim = f.insightface_sim ? f.insightface_sim.toFixed(3) : 'N/A';
        return `<div class="frame-item ${cssClass}">Frame ${f.frame}: ${status} (similarity: ${sim})</div>`;
    }).join('');
    document.getElementById('verify-frames').innerHTML = frameHtml;
    
    if (data.messages && data.messages.length > 0) {
        const msgHtml = data.messages.map(m => `<div>Frame ${m.frame}: ${m.error}</div>`).join('');
        document.getElementById('verify-errors').innerHTML = msgHtml;
        document.getElementById('verify-errors-details').style.display = 'block';
    }
    
    resultDiv.classList.remove('error');
    resultDiv.classList.add('success');
}

function showVerifyError(data) {
    const resultDiv = document.getElementById('verify-result');
    
    document.getElementById('verify-title').textContent = '❌ ' + data.message;
    document.getElementById('verify-message').textContent = `Reason: ${data.reason}`;
    
    // Stats (if available)
    if (data.stats) {
        document.getElementById('stat-verification').textContent = data.stats.verification_rate?.toFixed(1) + '%' || 'N/A';
        document.getElementById('stat-spoof').textContent = data.stats.spoof_rate?.toFixed(1) + '%' || 'N/A';
        document.getElementById('stat-real').textContent = data.stats.real_rate?.toFixed(1) + '%' || 'N/A';
        document.getElementById('stat-frames').textContent = data.stats.processed_frames || 'N/A';
        document.getElementById('verify-stats').classList.remove('hidden');
    }
    
    // Detailed stats
    if (data.stats) {
        const details = formatJson(data.stats);
        document.getElementById('verify-details').innerHTML = `<pre>${details}</pre>`;
    }
    
    // Messages
    if (data.messages && data.messages.length > 0) {
        const msgHtml = data.messages.map(m => `<div>Frame ${m.frame}: ${m.error}</div>`).join('');
        document.getElementById('verify-errors').innerHTML = msgHtml;
        document.getElementById('verify-errors-details').style.display = 'block';
    } else {
        document.getElementById('verify-errors-details').style.display = 'none';
    }
    
    document.getElementById('verify-frames').innerHTML = '';
    
    resultDiv.classList.remove('success');
    resultDiv.classList.add('error');
}

function formatJson(obj) {
    return JSON.stringify(obj, null, 2);
}
