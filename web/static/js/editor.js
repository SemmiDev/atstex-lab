// ── PDF.js setup ─────────────────────────────────────────────
    import * as pdfjsLib from 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/4.4.168/pdf.min.mjs';
    pdfjsLib.GlobalWorkerOptions.workerSrc = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/4.4.168/pdf.worker.min.mjs';

    // ── CodeMirror 6 setup ───────────────────────────────────────
    import { EditorView, basicSetup } from "https://esm.sh/codemirror";
    import { EditorState } from "https://esm.sh/@codemirror/state";
    import { keymap } from "https://esm.sh/@codemirror/view";
    import { standardKeymap, indentWithTab } from "https://esm.sh/@codemirror/commands";

    // We will use a generic stream-syntax for LaTeX since there isn't a robust official Lezer grammar in core yet
    import { StreamLanguage } from "https://esm.sh/@codemirror/language";
    import { stex } from "https://esm.sh/@codemirror/legacy-modes/mode/stex";
    import { autocompletion, snippetCompletion } from "https://esm.sh/@codemirror/autocomplete";

    // ── State ─────────────────────────────────────────────────────
    let pdfBytes = null;
    let pdfDoc = null;
    let zoom = 1.0;
    let currentEngine = 'pdflatex';

    // ── DOM refs ──────────────────────────────────────────────────
    const editorContainer = document.getElementById('editor-container');
    const btnCompile = document.getElementById('btn-compile');
    const btnDownload = document.getElementById('btn-download');
    const engineSel = document.getElementById('engine');
    const templateSel = document.getElementById('template');
    const previewScroll = document.getElementById('preview-scroll');
    const previewEmpty = document.getElementById('preview-empty');
    const logPanel = document.getElementById('log-panel');
    const logOutput = document.getElementById('log-output');
    const logBadge = document.getElementById('log-badge');
    const logToggle = document.getElementById('log-toggle');
    const statusDot = document.getElementById('status-dot');
    const statusText = document.getElementById('status-text');
    const statusElapsed = document.getElementById('status-elapsed');
    const statusPages = document.getElementById('status-pages');
    const statusSep2 = document.getElementById('status-sep2');
    const zoomVal = document.getElementById('zoom-val');
    const zoomIn = document.getElementById('zoom-in');
    const zoomOut = document.getElementById('zoom-out');
    const zoomFit = document.getElementById('zoom-fit');

    // ── Example document ──────────────────────────────────────────
    const EXAMPLE = String.raw`\documentclass[12pt,a4paper]{article}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{geometry}
\usepackage{amsmath,amssymb,amsthm}
\usepackage{hyperref}
\usepackage{booktabs}
\usepackage{graphicx}
\usepackage{xcolor}
\usepackage{enumitem}
\usepackage{fancyhdr}
\geometry{margin=2.5cm}
\pagestyle{fancy}
\fancyhf{}
\rhead{\textcolor{gray}{\small LaTeXPad Demo}}
\lhead{\textcolor{gray}{\small \today}}
\cfoot{\thepage}

\title{\textbf{LaTeXPad Demonstration}\\
  \large A Complex Document Example}
\author{LaTeXPad User}
\date{\today}

\newtheorem{theorem}{Theorem}
\newtheorem{lemma}[theorem]{Lemma}

\begin{document}
\maketitle
\tableofcontents
\newpage

\section{Introduction}
This document demonstrates that \textbf{LaTeXPad} can compile complex
\LaTeX{} documents including mathematics, tables, and theorems.

\section{Mathematics}

\subsection{Inline and Display Math}
The quadratic formula $x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}$ solves
any quadratic equation $ax^2 + bx + c = 0$.

Maxwell's equations in differential form:
\begin{align}
  \nabla \cdot \mathbf{E}  &= \frac{\rho}{\varepsilon_0} \\
  \nabla \cdot \mathbf{B}  &= 0 \\
  \nabla \times \mathbf{E} &= -\frac{\partial \mathbf{B}}{\partial t} \\
  \nabla \times \mathbf{B} &= \mu_0 \left(\mathbf{J} +
    \varepsilon_0 \frac{\partial \mathbf{E}}{\partial t}\right)
\end{align}

\subsection{A Theorem}
\begin{theorem}[Pythagorean Theorem]
  For a right triangle with legs $a$, $b$ and hypotenuse $c$:
  \[ a^2 + b^2 = c^2 \]
\end{theorem}

\begin{proof}
  This classical result follows from the geometric construction of squares
  on each side of the triangle and the equality of areas.
\end{proof}

\section{Tables}
\begin{table}[h!]
  \centering
  \caption{Comparison of LaTeX Engines}
  \begin{tabular}{@{}llll@{}}
    \toprule
    Engine     & Unicode & OpenType & Speed  \\ \midrule
    pdflatex   & limited & no       & fast   \\
    xelatex    & full    & yes      & medium \\
    lualatex   & full    & yes      & slow   \\ \bottomrule
  \end{tabular}
\end{table}

\section{Lists}
\begin{itemize}[leftmargin=*]
  \item \textbf{pdflatex} — most compatible, widest package support
  \item \textbf{xelatex} — native Unicode and system font access
  \item \textbf{lualatex} — scriptable via embedded Lua
\end{itemize}

\section{Conclusion}
\LaTeX{} remains the gold standard for typesetting scientific and
mathematical documents. \textit{LaTeXPad} brings this power directly
to your browser.

\end{document}`;

    // ── CodeMirror Initialization ─────────────────────────────────
    const cmTheme = EditorView.theme({
      "&": { backgroundColor: "var(--bg)", color: "var(--text)" },
      ".cm-gutters": { backgroundColor: "var(--surface)", color: "var(--muted)", borderRight: "3px solid var(--border)", fontWeight: "bold" },
      "&.cm-focused .cm-cursor": { borderLeftColor: "var(--accent)", borderLeftWidth: "3px" },
      "&.cm-focused .cm-selectionBackground, ::selection": { backgroundColor: "rgba(255, 71, 148, 0.2)" }
    }, { dark: false });

    // ── Autocomplete dictionary ───────────────────────────────────
    const latexCompletions = [
      snippetCompletion('\\begin{${environment}}\n\t${}\n\\end{${environment}}', { label: 'begin', info: 'Begin environment' }),
      snippetCompletion('\\section{${title}}', { label: 'section', info: 'Section' }),
      snippetCompletion('\\subsection{${title}}', { label: 'subsection', info: 'Subsection' }),
      snippetCompletion('\\textbf{${text}}', { label: 'textbf', info: 'Bold text' }),
      snippetCompletion('\\textit{${text}}', { label: 'textit', info: 'Italic text' }),
      snippetCompletion('\\frac{${num}}{${den}}', { label: 'frac', info: 'Fraction' }),
      snippetCompletion('\\sum_{${lower}}^{${upper}}', { label: 'sum', info: 'Summation' }),
      snippetCompletion('\\int_{${lower}}^{${upper}}', { label: 'int', info: 'Integral' }),
      { label: '\\alpha', type: 'keyword' },
      { label: '\\beta', type: 'keyword' },
      { label: '\\gamma', type: 'keyword' },
      { label: '\\Delta', type: 'keyword' },
      { label: '\\theta', type: 'keyword' },
      { label: '\\infty', type: 'keyword' },
      { label: '\\rightarrow', type: 'keyword' },
      { label: '\\maketitle', type: 'keyword' },
      { label: '\\tableofcontents', type: 'keyword' },
      { label: '\\newpage', type: 'keyword' },
      { label: '\\usepackage{${package}}', type: 'keyword' },
    ];

    function latexCompletionSource(context) {
      let word = context.matchBefore(/\\[a-zA-Z]*/)
      if (!word) return null
      if (word.from == word.to && !context.explicit) return null
      return {
        from: word.from,
        options: latexCompletions
      }
    }

    const editorView = new EditorView({
      state: EditorState.create({
        doc: EXAMPLE,
        extensions: [
          basicSetup,
          StreamLanguage.define(stex),
          cmTheme,
          autocompletion({ override: [latexCompletionSource] }),
          keymap.of([
            indentWithTab,
            {
              key: "Mod-Enter",
              run: () => {
                if (!btnCompile.disabled) compile();
                return true;
              }
            }
          ])
        ]
      }),
      parent: editorContainer
    });

    const compileOverlay = document.getElementById('compile-overlay');
    const compileStep = document.getElementById('compile-step');

    // ── Compile ───────────────────────────────────────────────────
    async function compile() {
      setStatus('compiling', 'Compiling…');
      btnCompile.disabled = true;
      btnCompile.classList.add('loading');
      compileOverlay.style.display = '';
      compileStep.textContent = 'Rendering template…';

      try {
        // If a real template is selected, re-render it with fresh biodata + page settings
        const selectedTemplate = templateSel.value;
        if (selectedTemplate && selectedTemplate !== 'example') {
          const cvDataStr = localStorage.getItem('cv_data') || '{}';
          let reqBody = {};
          try { reqBody = JSON.parse(cvDataStr); } catch (_) {}
          // Include page settings if any are saved
          try {
            const savedSettings = JSON.parse(localStorage.getItem('page_settings') || 'null');
            if (savedSettings) reqBody.settings = savedSettings;
          } catch (_) {}
          const renderRes = await fetch(`/api/templates/${selectedTemplate}/render`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(reqBody)
          });
          if (renderRes.ok) {
            const content = await renderRes.text();
            editorView.dispatch({
              changes: { from: 0, to: editorView.state.doc.length, insert: content }
            });
          }
        }

        const source = editorView.state.doc.toString();
        if (!source.trim()) {
          setStatus('error', 'Editor is empty');
          return;
        }

        compileStep.textContent = 'Running LaTeX compiler…';
        const res = await fetch('/compile', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ source, engine: engineSel.value }),
        });

        if (res.ok && res.headers.get('Content-Type')?.startsWith('application/pdf')) {
          const elapsed = res.headers.get('X-Latex-Elapsed') || '';
          const rawLog = res.headers.get('X-Latex-Log') || '';
          const buf = await res.arrayBuffer();
          pdfBytes = new Uint8Array(buf);
          currentEngine = engineSel.value;

          showLog(rawLog, true);
          await renderPDF(pdfBytes);
          setStatus('success', 'Compiled successfully');
          statusElapsed.textContent = elapsed ? `${elapsed}` : '';
          btnDownload.disabled = false;
        } else {
          const data = await res.json();
          showLog(data.log || data.error || 'Unknown error', false);
          setStatus('error', data.error || 'Compilation failed');
        }
      } catch (err) {
        showLog(String(err), false);
        setStatus('error', 'Network error');
      } finally {
        btnCompile.disabled = false;
        btnCompile.classList.remove('loading');
        compileOverlay.style.display = 'none';
      }
    }

    // ── Render PDF pages ──────────────────────────────────────────
    async function renderPDF(bytes) {
      pdfDoc = await pdfjsLib.getDocument({ data: bytes.slice() }).promise;
      await renderAllPages();

      statusSep2.style.display = '';
      statusPages.textContent = `${pdfDoc.numPages} page${pdfDoc.numPages !== 1 ? 's' : ''}`;
    }

    async function renderAllPages() {
      if (!pdfDoc) return;

      previewEmpty.style.display = 'none';

      // Remove old canvases
      previewScroll.querySelectorAll('canvas.pdf-page').forEach(c => c.remove());

      for (let i = 1; i <= pdfDoc.numPages; i++) {
        const page = await pdfDoc.getPage(i);
        const viewport = page.getViewport({ scale: zoom * devicePixelRatio });

        const canvas = document.createElement('canvas');
        canvas.className = 'pdf-page';
        canvas.width = viewport.width;
        canvas.height = viewport.height;
        canvas.style.width = `${viewport.width / devicePixelRatio}px`;
        canvas.style.height = `${viewport.height / devicePixelRatio}px`;

        previewScroll.appendChild(canvas);

        const ctx = canvas.getContext('2d');
        await page.render({ canvasContext: ctx, viewport }).promise;
      }
    }

    // ── Zoom ──────────────────────────────────────────────────────
    function setZoom(z) {
      zoom = Math.max(0.4, Math.min(3.0, z));
      zoomVal.textContent = `${Math.round(zoom * 100)}%`;
      renderAllPages();
    }

    zoomIn.addEventListener('click', () => setZoom(zoom + 0.2));
    zoomOut.addEventListener('click', () => setZoom(zoom - 0.2));
    zoomFit.addEventListener('click', () => {
      const paneW = previewScroll.clientWidth - 48;
      // Default A4 at scale 1 renders at ~595px wide
      const naturalW = 595;
      setZoom(paneW / naturalW);
    });

    // ── Download PDF ──────────────────────────────────────────────
    // ── Download PDF + Trigger Donation Modal ─────────────────────
    const donationModal = document.getElementById('donation-modal');
    const closeDonationModal = document.getElementById('close-donation-modal');

    btnDownload.addEventListener('click', () => {
      if (!pdfBytes) return;
      const blob = new Blob([pdfBytes], { type: 'application/pdf' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.style.display = 'none';
      a.href = url;
      a.download = 'document.pdf';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      setTimeout(() => URL.revokeObjectURL(url), 1000);

      // Show the Saweria popup right after the file triggers
      donationModal.classList.add('active');
    });

    // ── Download LaTeX source ─────────────────────────────────────
    const btnDownloadTex = document.getElementById('btn-download-tex');
    btnDownloadTex.addEventListener('click', () => {
      const source = editorView.state.doc.toString();
      if (!source.trim()) return;
      const blob = new Blob([source], { type: 'application/x-latex' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.style.display = 'none';
      a.href = url;
      a.download = 'document.tex';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    });

    closeDonationModal.addEventListener('click', () => {
      donationModal.classList.remove('active');
    });

    // ── Log panel ─────────────────────────────────────────────────
    function showLog(raw, success) {
      if (!raw) return;

      logOutput.innerHTML = raw
        .split('\n')
        .map(line => {
          const esc = line.replace(/&/g, '&amp;').replace(/</g, '&lt;');
          if (/error|fatal|undefined control|!/.test(line.toLowerCase()))
            return `<span class="err-line">${esc}</span>`;
          if (/warning/i.test(line))
            return `<span class="warn-line">${esc}</span>`;
          return esc;
        })
        .join('\n');

      logBadge.textContent = success ? 'OK' : 'ERR';
      logBadge.className = success ? 'badge ok' : 'badge';
      logBadge.style.display = '';
      logPanel.classList.add('open');
      logOutput.scrollTop = logOutput.scrollHeight;
    }

    logToggle.addEventListener('click', () => {
      logPanel.classList.toggle('open');
    });

    // ── Status bar ────────────────────────────────────────────────
    function setStatus(state, msg) {
      statusDot.className = `status-dot ${state}`;
      statusText.textContent = msg;
      if (state !== 'compiling') return;
      statusElapsed.textContent = '';
      statusPages.textContent = '';
      statusSep2.style.display = 'none';
    }

    // ── Keyboard shortcut: Ctrl+Enter / Cmd+Enter ─────────────────
    document.addEventListener('keydown', e => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault();
        if (!btnCompile.disabled) compile();
      }
    });

    btnCompile.addEventListener('click', compile);

    // ── Splitter drag ─────────────────────────────────────────────
    const splitter = document.getElementById('splitter');
    const editorPane = document.getElementById('editor-pane');
    const biodataFrameRef = document.getElementById('biodata-frame');
    let dragging = false;

    splitter.addEventListener('mousedown', e => {
      dragging = true;
      splitter.classList.add('dragging');
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
      // Prevent iframe from stealing mouse events during drag
      if (biodataFrameRef) biodataFrameRef.style.pointerEvents = 'none';
      e.preventDefault();
    });

    document.addEventListener('mousemove', e => {
      if (!dragging) return;
      const workspace = document.querySelector('.workspace');
      const rect = workspace.getBoundingClientRect();
      const x = Math.max(280, Math.min(e.clientX - rect.left, rect.width - 280));
      editorPane.style.width = x + 'px';
    });

    document.addEventListener('mouseup', () => {
      if (!dragging) return;
      dragging = false;
      splitter.classList.remove('dragging');
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
      // Restore iframe pointer events
      if (biodataFrameRef) biodataFrameRef.style.pointerEvents = '';
    });

    // ── Template Selector ─────────────────────────────────────────
    async function fetchTemplates() {
      try {
        const res = await fetch('/api/templates');
        if (!res.ok) return;
        const tpls = await res.json();
        for (const tpl of tpls) {
          const opt = document.createElement('option');
          opt.value = tpl.name;
          opt.textContent = tpl.description || tpl.name;
          templateSel.appendChild(opt);
        }
      } catch (err) {
        console.error("Failed to fetch templates:", err);
      }
    }
    fetchTemplates();

    templateSel.addEventListener('change', async (e) => {
      const name = e.target.value;
      if (!name) {
        editorView.dispatch({
          changes: { from: 0, to: editorView.state.doc.length, insert: "" }
        });
        editorView.focus();
        return;
      }

      if (name === "example") {
        editorView.dispatch({
          changes: { from: 0, to: editorView.state.doc.length, insert: EXAMPLE }
        });
        editorView.focus();
        if (!btnCompile.disabled) compile();
        return;
      }

      try {
        const cvDataStr = localStorage.getItem('cv_data') || '{}';
        let reqBody = {};
        try { reqBody = JSON.parse(cvDataStr); } catch (_) {}
        // Include page settings if any are saved
        try {
          const savedSettings = JSON.parse(localStorage.getItem('page_settings') || 'null');
          if (savedSettings) reqBody.settings = savedSettings;
        } catch (_) {}
        const res = await fetch(`/api/templates/${name}/render`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(reqBody)
        });
        if (!res.ok) throw new Error('Failed to load template');
        const content = await res.text();
        editorView.dispatch({
          changes: { from: 0, to: editorView.state.doc.length, insert: content }
        });
        editorView.focus();

        // Auto compile
        if (!btnCompile.disabled) compile();
      } catch (err) {
        console.error(err);
        setStatus('error', 'Failed to load template');
      }
    });

    // ── Page Settings: Apply & Reset ─────────────────────────────
    const btnApplySettings = document.getElementById('btn-apply-settings');
    const btnResetSettings = document.getElementById('btn-reset-settings');

    function getPageSettings() {
      return {
        documentClass: document.getElementById('ps-doc-class').value,
        paperSize: document.getElementById('ps-paper-size').value,
        fontSize: document.getElementById('ps-font-size').value,
        fontFamily: document.getElementById('ps-font-family').value,
        marginTop: document.getElementById('ps-margin-top').value,
        marginBottom: document.getElementById('ps-margin-bottom').value,
        marginLeft: document.getElementById('ps-margin-left').value,
        marginRight: document.getElementById('ps-margin-right').value,
        lineSpacing: parseFloat(document.getElementById('ps-line-spacing').value) || 1.0,
        alignment: document.getElementById('ps-alignment').value,
        headerText: document.getElementById('ps-header-text').value,
        footerText: document.getElementById('ps-footer-text').value,
      };
    }

    function setPageSettings(s) {
      if (!s) return;
      if (s.documentClass) document.getElementById('ps-doc-class').value = s.documentClass;
      if (s.paperSize) document.getElementById('ps-paper-size').value = s.paperSize;
      if (s.fontSize) document.getElementById('ps-font-size').value = s.fontSize;
      if (s.fontFamily) document.getElementById('ps-font-family').value = s.fontFamily;
      if (s.marginTop) document.getElementById('ps-margin-top').value = s.marginTop;
      if (s.marginBottom) document.getElementById('ps-margin-bottom').value = s.marginBottom;
      if (s.marginLeft) document.getElementById('ps-margin-left').value = s.marginLeft;
      if (s.marginRight) document.getElementById('ps-margin-right').value = s.marginRight;
      if (s.lineSpacing) document.getElementById('ps-line-spacing').value = String(s.lineSpacing);
      if (s.alignment) document.getElementById('ps-alignment').value = s.alignment;
      if (s.headerText !== undefined) document.getElementById('ps-header-text').value = s.headerText;
      if (s.footerText !== undefined) document.getElementById('ps-footer-text').value = s.footerText;
    }

    // Restore saved settings from localStorage
    try {
      const saved = JSON.parse(localStorage.getItem('page_settings') || 'null');
      if (saved) setPageSettings(saved);
    } catch (_) { /* ignore parse errors */ }

    if (btnApplySettings) {
      btnApplySettings.addEventListener('click', async () => {
        const selectedTemplate = templateSel.value;
        if (!selectedTemplate || selectedTemplate === 'example') {
          setStatus('error', 'Select a CV template first (not Blank or Example)');
          return;
        }

        const settings = getPageSettings();
        const cvDataStr = localStorage.getItem('cv_data') || '{}';
        let cvData = {};
        try { cvData = JSON.parse(cvDataStr); } catch (_) {}

        // Persist settings to localStorage
        localStorage.setItem('page_settings', JSON.stringify(settings));

        setStatus('compiling', 'Applying page settings…');
        btnApplySettings.disabled = true;

        try {
          const res = await fetch('/api/page-settings/apply', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ template: selectedTemplate, cvData, settings }),
          });

          if (!res.ok) {
            const err = await res.json();
            setStatus('error', err.error || 'Failed to apply settings');
            return;
          }

          const modifiedSource = await res.text();
          editorView.dispatch({
            changes: { from: 0, to: editorView.state.doc.length, insert: modifiedSource }
          });

          setStatus('success', 'Page settings applied');

          // Auto compile after applying settings
          if (!btnCompile.disabled) compile();
        } catch (err) {
          console.error(err);
          setStatus('error', 'Failed to apply page settings');
        } finally {
          btnApplySettings.disabled = false;
        }
      });
    }

    if (btnResetSettings) {
      btnResetSettings.addEventListener('click', () => {
        const defaults = {
          documentClass: 'article',
          paperSize: 'a4paper',
          fontSize: '10pt',
          fontFamily: 'default',
          marginTop: '0.5in',
          marginBottom: '0.5in',
          marginLeft: '0.7in',
          marginRight: '0.7in',
          lineSpacing: 1,
          alignment: 'justify',
          headerText: '',
          footerText: '',
        };
        setPageSettings(defaults);
        localStorage.removeItem('page_settings');
      });
    }
