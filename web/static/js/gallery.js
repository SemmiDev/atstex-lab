// ── PDF.js setup ─────────────────────────────────────────────
import * as pdfjsLib from "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/4.4.168/pdf.min.mjs";
pdfjsLib.GlobalWorkerOptions.workerSrc =
  "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/4.4.168/pdf.worker.min.mjs";

// ── User-scoped localStorage keys ─────────────────────────────
function _galleryUid() {
  return document.body.dataset.userId || "anonymous";
}
function _glsKey(base) {
  return `${base}_${_galleryUid()}`;
}

// ── DOM refs ──────────────────────────────────────────────────
const galleryGrid = document.getElementById("gallery-grid");
const btnGenerate = document.getElementById("btn-generate");
const statusDot = document.getElementById("gallery-status-dot");
const statusText = document.getElementById("gallery-status-text");
const statusElapsed = document.getElementById("gallery-status-elapsed");

// ── Template metadata ─────────────────────────────────────────
const TEMPLATE_COLORS = ["#ff4794", "#475eff", "#5eeb8f", "#ffdb33", "#ff8c3a", "#b066ff"];
const TEMPLATE_EMOJIS = ["🌊", "📐", "🪸", "🌊", "🌊", "🌊"];

// ── Status helpers ────────────────────────────────────────────
function setStatus(state, msg) {
  statusDot.className = `status-dot ${state}`;
  statusText.textContent = msg;
  if (state === "compiling") statusElapsed.textContent = "";
}

// ── Build card skeleton ───────────────────────────────────────
function createCardSkeleton(name, idx) {
  const card = document.createElement("div");
  card.className = "gallery-card";
  card.id = `card-${name}`;

  const color = TEMPLATE_COLORS[idx % TEMPLATE_COLORS.length];

  card.innerHTML = `
    <div class="gallery-card-header">
      <span class="gallery-card-title">
        <span class="dot" style="background:${color}"></span>
        ${name}
      </span>
      <span class="gallery-card-badge">.tex</span>
    </div>
    <div class="gallery-card-preview" id="preview-${name}">
      <div class="gallery-skeleton">
        <span class="skeleton-bounce">⚙️</span>
        <span class="skeleton-text">Memproses Dokumen…</span>
      </div>
    </div>
    <div class="gallery-card-footer">
      <button class="btn-use" disabled data-template="${name}">
        <i class="ph-bold ph-arrow-right"></i> Gunakan Template
      </button>
    </div>
  `;
  return card;
}

// ── Render first page of PDF to canvas ────────────────────────
async function renderPDFToCanvas(base64DataUrl, container) {
  // Strip data URL prefix to get raw base64
  const raw = base64DataUrl.split(",")[1];
  const binary = atob(raw);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);

  const pdf = await pdfjsLib.getDocument({ data: bytes }).promise;
  const page = await pdf.getPage(1);

  // Calculate scale to fit within the card preview area
  const containerWidth = container.clientWidth || 320;
  const viewport = page.getViewport({ scale: 1 });
  const scale = (containerWidth / viewport.width) * devicePixelRatio;
  const scaledViewport = page.getViewport({ scale });

  const canvas = document.createElement("canvas");
  canvas.width = scaledViewport.width;
  canvas.height = scaledViewport.height;
  canvas.style.width = `${scaledViewport.width / devicePixelRatio}px`;
  canvas.style.height = `${scaledViewport.height / devicePixelRatio}px`;

  const ctx = canvas.getContext("2d");
  await page.render({ canvasContext: ctx, viewport: scaledViewport }).promise;

  container.innerHTML = "";
  container.appendChild(canvas);
}

// ── Show error on a card ──────────────────────────────────────
function showCardError(container, msg) {
  container.innerHTML = `
    <div class="gallery-card-error">
      <span class="error-icon">⚠️</span>
      <span class="error-msg">${msg}</span>
    </div>
  `;
}

// ── Generate all previews ─────────────────────────────────────
async function generateAll() {
  const cvDataStr = localStorage.getItem(_glsKey("cv_data")) || "{}";
  let reqBody = {};
  try {
    reqBody = JSON.parse(cvDataStr);
  } catch (_) {}

  // Include page settings
  try {
    const savedSettings = JSON.parse(localStorage.getItem(_glsKey("page_settings")) || "null");
    if (savedSettings) reqBody.settings = savedSettings;
  } catch (_) {}

  // Check if user has any meaningful data
  const hasData = reqBody.personal && (reqBody.personal.name || reqBody.personal.email);

  setStatus("compiling", "Generating previews…");
  btnGenerate.disabled = true;
  btnGenerate.classList.add("loading");

  const start = Date.now();

  try {
    // First fetch template list to build the grid
    const tplRes = await fetch("/api/templates");
    if (!tplRes.ok) throw new Error("Failed to fetch template list");
    const templates = await tplRes.json();

    // Build skeleton cards
    galleryGrid.innerHTML = "";
    templates.forEach((tpl, idx) => {
      galleryGrid.appendChild(createCardSkeleton(tpl.name, idx));
    });

    if (!hasData) {
      // Show a friendly message on each card
      templates.forEach((tpl) => {
        const container = document.getElementById(`preview-${tpl.name}`);
        container.innerHTML = `
          <div class="gallery-card-error">
            <span class="error-icon">📝</span>
            <span class="error-msg">Isi biodata Anda terlebih dahulu di halaman <strong>Biodata</strong>, lalu kembali ke sini.</span>
          </div>
        `;
      });
      setStatus("error", "No biodata found — fill in your data first");
      return;
    }

    // Call batch compile
    const res = await fetch("/api/gallery/compile", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(reqBody),
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error((err.detail || err.error) || "Compilation failed");
    }

    const results = await res.json();
    let successCount = 0;

    for (const result of results) {
      const container = document.getElementById(`preview-${result.name}`);
      const btn = document.querySelector(`button[data-template="${result.name}"]`);
      if (!container) continue;

      if ((result.detail || result.error)) {
        showCardError(container, (result.detail || result.error).substring(0, 200));
      } else if (result.pdf_base64) {
        try {
          await renderPDFToCanvas(result.pdf_base64, container);
          if (btn) btn.disabled = false;
          successCount++;
        } catch (e) {
          showCardError(container, "Failed to render PDF preview");
        }
      }
    }

    const elapsed = ((Date.now() - start) / 1000).toFixed(1);
    statusElapsed.textContent = `${elapsed}s`;
    setStatus("success", `${successCount}/${results.length} templates rendered`);
  } catch (err) {
    setStatus("error", err.message || "Network error");
  } finally {
    btnGenerate.disabled = false;
    btnGenerate.classList.remove("loading");
  }
}

// ── "Use Template" click handler ──────────────────────────────
galleryGrid.addEventListener("click", (e) => {
  const btn = e.target.closest(".btn-use");
  if (!btn || btn.disabled) return;
  const templateName = btn.dataset.template;
  if (!templateName) return;

  // Store the selected template for the editor to pick up
  localStorage.setItem(_glsKey("gallery_selected_template"), templateName);
  window.location.href = "/editor";
});

// ── Auto-generate on load ─────────────────────────────────────
btnGenerate.addEventListener("click", generateAll);
generateAll();
