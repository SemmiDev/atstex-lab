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

    // ── Compile ───────────────────────────────────────────────────
    async function compile() {
      const source = editorView.state.doc.toString();
      if (!source.trim()) return;

      setStatus('compiling', 'Compiling…');
      btnCompile.disabled = true;
      btnCompile.classList.add('loading');

      try {
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
    let dragging = false;

    splitter.addEventListener('mousedown', e => {
      dragging = true;
      splitter.classList.add('dragging');
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
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
      dragging = false;
      splitter.classList.remove('dragging');
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
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
        const res = await fetch(`/api/templates/${name}/render`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: cvDataStr
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
