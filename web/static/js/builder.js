let currentTemplateId = null;
let currentTemplateName = "Template Kustom 1";

document.addEventListener('DOMContentLoaded', () => {
    const components = document.querySelectorAll('.component-item');
    const canvas = document.getElementById('canvas-area');
    const placeholder = document.getElementById('canvas-placeholder');

    // Drag start
    components.forEach(comp => {
        comp.addEventListener('dragstart', (e) => {
            e.dataTransfer.setData('text/plain', comp.dataset.type);
            e.dataTransfer.effectAllowed = 'copy';
            comp.style.opacity = '0.5';
        });

        comp.addEventListener('dragend', () => {
            comp.style.opacity = '1';
        });
    });

    // Drag over canvas
    canvas.addEventListener('dragover', (e) => {
        e.preventDefault();
        e.dataTransfer.dropEffect = 'copy';
        canvas.style.backgroundColor = 'var(--accent3)';
    });

    canvas.addEventListener('dragleave', () => {
        canvas.style.backgroundColor = '';
    });

    // Drop on canvas
    canvas.addEventListener('drop', (e) => {
        e.preventDefault();
        canvas.style.backgroundColor = '';
        
        const type = e.dataTransfer.getData('text/plain');
        if (!type) return;

        if (placeholder) {
            placeholder.style.display = 'none';
        }

        addBlockToCanvas(type);
    });
});

function getBlockLabel(type) {
    switch (type) {
        case 'header': return 'Header / Kontak';
        case 'summary': return 'Ringkasan';
        case 'experience': return 'Pengalaman Kerja';
        case 'education': return 'Pendidikan';
        case 'skills': return 'Keahlian';
        case 'projects': return 'Proyek';
        case 'certifications': return 'Sertifikasi';
        case 'languages': return 'Bahasa';
        default: return 'Blok';
    }
}

function addBlockToCanvas(type) {
    const canvas = document.getElementById('canvas-area');
    const block = document.createElement('div');
    block.className = 'dropped-block';
    block.dataset.type = type;
    
    block.innerHTML = `
        <div class="dropped-block-title">${getBlockLabel(type)}</div>
        <div class="font-mono text-xs text-muted">Komponen dinamis akan dirender di sini.</div>
        <div class="block-actions">
            <button onclick="moveBlockUp(this)"><i class="ph-bold ph-arrow-up"></i></button>
            <button onclick="moveBlockDown(this)"><i class="ph-bold ph-arrow-down"></i></button>
            <button onclick="this.parentElement.parentElement.remove(); checkEmptyCanvas();" class="text-error"><i class="ph-bold ph-trash"></i></button>
        </div>
    `;
    
    canvas.appendChild(block);
}

function moveBlockUp(btn) {
    const block = btn.closest('.dropped-block');
    if (block.previousElementSibling && block.previousElementSibling.id !== 'canvas-placeholder') {
        block.parentNode.insertBefore(block, block.previousElementSibling);
    }
}

function moveBlockDown(btn) {
    const block = btn.closest('.dropped-block');
    if (block.nextElementSibling) {
        block.parentNode.insertBefore(block.nextElementSibling, block);
    }
}

function checkEmptyCanvas() {
    const canvas = document.getElementById('canvas-area');
    const blocks = canvas.querySelectorAll('.dropped-block');
    if (blocks.length === 0) {
        document.getElementById('canvas-placeholder').style.display = 'block';
    }
}

function clearCanvas() {
    if (!confirm('Apakah kamu yakin ingin membersihkan kanvas?')) return;
    const canvas = document.getElementById('canvas-area');
    const blocks = canvas.querySelectorAll('.dropped-block');
    blocks.forEach(b => b.remove());
    document.getElementById('canvas-placeholder').style.display = 'block';
}

function getTemplateConfig() {
    const blocks = document.querySelectorAll('.dropped-block');
    const layout = Array.from(blocks).map(b => b.dataset.type);
    
    return {
        theme_color: document.getElementById('theme-color').value,
        columns: parseInt(document.getElementById('layout-columns').value) || 1,
        layout: layout
    };
}

async function previewPDF() {
    const config = getTemplateConfig();
    if (config.layout.length === 0) {
        alert('Kanvas masih kosong!');
        return;
    }

    document.getElementById('preview-placeholder').innerHTML = '<i class="ph-bold ph-spinner ph-spin text-2xl"></i><p class="mt-2">Mempersiapkan PDF...</p>';
    document.getElementById('live-pdf-preview').style.display = 'none';

    try {
        const res = await fetch('/api/builder/preview', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });

        if (!res.ok) {
            const err = await res.text();
            throw new Error(err);
        }

        const blob = await res.blob();
        const url = URL.createObjectURL(blob);
        
        document.getElementById('live-pdf-preview').src = url;
        document.getElementById('live-pdf-preview').style.display = 'block';
        document.getElementById('preview-placeholder').style.display = 'none';
        
    } catch (err) {
        console.error(err);
        document.getElementById('preview-placeholder').innerHTML = `<p class="text-error">Gagal memuat pratinjau</p><p class="text-xs text-muted mt-2">${err.message}</p>`;
    }
}

async function saveTemplate() {
    const config = getTemplateConfig();
    if (config.layout.length === 0) {
        alert('Kanvas masih kosong!');
        return;
    }

    let payload = {
        config: config
    };

    if (currentTemplateId) {
        payload.id = currentTemplateId;
        payload.name = currentTemplateName;
    } else {
        const name = prompt("Beri nama template kustom ini:", currentTemplateName);
        if (!name) return;
        payload.name = name;
    }

    try {
        const res = await fetch('/api/builder/save', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (res.ok) {
            alert('Template berhasil disimpan!');
        } else {
            const err = await res.text();
            alert('Gagal menyimpan: ' + err);
        }
    } catch (e) {
        alert('Terjadi kesalahan koneksi.');
    }
}

// Modal Logic for Loading Templates
function openLoadModal() {
    const modal = document.getElementById('load-modal');
    modal.classList.remove('hidden');
    modal.classList.add('flex');
    loadTemplatesData();
}

function closeLoadModal() {
    const modal = document.getElementById('load-modal');
    modal.classList.add('hidden');
    modal.classList.remove('flex');
}

async function loadTemplatesData() {
    const content = document.getElementById('load-modal-content');
    content.innerHTML = '<div class="text-center font-mono text-sm text-muted py-4"><i class="ph-spin ph-spinner text-2xl"></i> Memuat...</div>';
    
    try {
        const res = await fetch('/api/builder/load');
        if (!res.ok) throw new Error('Gagal mengambil data');
        const templates = await res.json();
        
        if (!templates || templates.length === 0) {
            content.innerHTML = '<div class="text-center font-mono text-sm text-muted py-4">Belum ada template yang disimpan.</div>';
            return;
        }

        content.innerHTML = templates.map(t => `
            <div class="p-3 border-2 border-border bg-bg flex justify-between items-center">
                <div>
                    <div class="font-bold">${t.name}</div>
                    <div class="text-xs text-muted font-mono mt-1">${new Date(t.created_at).toLocaleDateString()}</div>
                </div>
                <button class="btn btn-outline text-xs px-3 py-1" onclick='applyTemplate(${JSON.stringify(t.config)}, "${t.id}", ${JSON.stringify(t.name)})'>
                    Pakai Template
                </button>
            </div>
        `).join('');
    } catch (err) {
        content.innerHTML = `<div class="text-center text-error font-mono text-sm py-4">${err.message}</div>`;
    }
}

function applyTemplate(config, id, name) {
    if (!confirm('Peringatan: ini akan mereset kanvas Anda saat ini. Lanjutkan?')) return;
    
    currentTemplateId = id || null;
    if (name) currentTemplateName = name;

    const canvas = document.getElementById('canvas-area');
    // Clear canvas
    canvas.querySelectorAll('.dropped-block').forEach(b => b.remove());
    
    // Set config
    if (config.theme_color) document.getElementById('theme-color').value = config.theme_color;
    if (config.columns) document.getElementById('layout-columns').value = config.columns;
    
    // Apply blocks
    if (config.layout && config.layout.length > 0) {
        const placeholder = document.getElementById('canvas-placeholder');
        if (placeholder) placeholder.style.display = 'none';

        config.layout.forEach(type => {
            addBlockToCanvas(type);
        });
    } else {
        document.getElementById('canvas-placeholder').style.display = 'block';
    }
    
    closeLoadModal();
}
