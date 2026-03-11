// i18n.js - Lightweight client-side localization

const dictionary = {
  en: {
    "nav.home": "Home",
    "nav.build": "Build",
    "hero.title.1": "Craft a Resume",
    "hero.title.2": "That Stands Out.",
    "hero.title.3": "Impress Every Recruiter.",
    "hero.subtitle":
      "ATSTEXLAB helps you build a professional, ATS-friendly resume that recruiters actually want to read. Just fill in your details, pick a clean template, and download your PDF.",
    "hero.btn.builder": "Build My Resume",
    "hero.btn.logout": "Logout",
    "hero.btn.signin": "Sign in with Google",
    "hero.trusted": "TRUSTED BY JOB SEEKERS, FRESH GRADS, AND CAREER SWITCHERS",
    "feature.1.title": "Instant PDF Export",
    "feature.1.desc":
      "Hit compile and get a perfectly formatted PDF in seconds. No software to install — everything runs in your browser.",
    "feature.3.title": "ATS-Friendly Templates",
    "feature.3.desc":
      "Every template is designed to pass Applicant Tracking Systems. Clean layout, clear sections, no hidden formatting tricks.",
    "feature.4.title": "Multiple CV Profiles",
    "feature.4.desc":
      'Create different resumes for different roles. One for "Back End Developer", another for "Data Engineer" — you decide.',
    "feature.5.title": "Import from PDF",
    "feature.5.desc":
      "Upload your old resume and let our AI instantly sort your experience into a new profile. Skip the manual typing.",
    "feature.6.title": "Get Instant AI Feedback",
    "feature.6.desc":
      "Get actionable feedback tailored by AI. Discover strengths, weaknesses, and a structured rating to help your resume stand out.",
    "feature.7.title": "AI Cover Letter Generator",
    "feature.7.desc":
      "Generate highly tailored cover letters combining your structured CV data and a specific job description.",
    "support.title": "Support the Project",
    "support.desc":
      "If you found this tool helpful, consider leaving a tip to keep the servers running!",
    "footer.copyright": "© 2026 ATSTEXLAB · Your Resume, Recruiter-Ready",
    "nav.ats_simulator": "ATS Simulator",
    "ats.title": "ATS Simulator",
    "ats.desc":
      "Score your CV against a Job Description. Paste text or upload an image to extract text.",
    "ats.form.profile": "CV Profile",
    "ats.form.language": "Simulation Language",
    "ats.form.jobdesc": "Job Description",
    "ats.form.upload": "Upload Image",
    "ats.btn.simulate": "Simulate ATS",
    "ats.loading": "Analyzing CV against Job Description…",
    "ats.result.score": "Match Score",
    "ats.result.missing": "Missing Keywords",
    "ats.result.recommendations": "Recommendations",
    "ats.history.title": "Simulation History",
  },
  id: {
    "nav.home": "Beranda",
    "nav.build": "Buat CV",
    "hero.title.1": "Buat CV",
    "hero.title.2": "Yang Menonjol.",
    "hero.title.3": "Bikin Rekruter Terkesan.",
    "hero.subtitle":
      "Bikin CV profesional yang ramah ATS tanpa ribet. Cukup isi profil kamu, pilih template yang rapi, dan langsung unduh hasilnya dalam format PDF.",
    "hero.btn.builder": "Buat CV Sekarang",
    "hero.btn.logout": "Keluar",
    "hero.btn.signin": "Masuk dengan Google",
    "hero.trusted":
      "DIPERCAYA OLEH PENCARI KERJA, LULUSAN BARU, DAN PROFESIONAL",
    "feature.1.title": "Ekspor PDF Instan",
    "feature.1.desc":
      "Klik export dan PDF kamu langsung jadi. Ngga perlu install aplikasi tambahan, semuanya jalan langsung di browser.",
    "feature.3.title": "Template Ramah ATS",
    "feature.3.desc":
      "Semua template dirancang khusus supaya gampang dibaca oleh sistem ATS perusahaan. Gak ada format aneh yang bikin CV kamu gagal screening.",
    "feature.4.title": "Banyak Profil CV",
    "feature.4.desc":
      'Buat CV berbeda untuk peran yang berbeda. Satu untuk "Back End Developer", satu lagi untuk "Data Engineer" — kamu yang tentukan.',
    "feature.5.title": "Import Langsung dari PDF",
    "feature.5.desc":
      "Punya CV lama? Upload aja. AI kami bakal otomatis mindahin isinya ke profil baru kamu. Bebas capek ngetik ulang.",
    "feature.6.title": "Dapatkan Masukan Instan AI",
    "feature.6.desc":
      "Cek seberapa bagus CV kamu dengan AI. Temukan kelebihan, bagian yang harus diperbaiki, dan skor keseluruhan biar CV makin stand out.",
    "feature.7.title": "Pembuat Cover Letter AI",
    "feature.7.desc":
      "Buat cover letter yang sangat sesuai dengan menggabungkan data CV kamu dan deskripsi pekerjaan incaranmu.",
    "support.title": "Dukung Proyek Ini",
    "support.desc":
      "Jika alat ini bermanfaat bagi Anda, pertimbangkan untuk memberikan tip agar server tetap menyala!",
    "footer.copyright": "© 2026 ATSTEXLAB · CV Anda, Siap Dipanggil",
    "nav.ats_simulator": "Simulator ATS",
    "ats.title": "Simulator ATS",
    "ats.desc":
      "Cocokkan CV Anda dengan Deskripsi Pekerjaan. Tempel teks atau unggah gambar untuk mengekstrak teks.",
    "ats.form.profile": "Profil CV",
    "ats.form.language": "Bahasa Simulasi",
    "ats.form.jobdesc": "Deskripsi Pekerjaan",
    "ats.form.upload": "Unggah Gambar (OCR)",
    "ats.btn.simulate": "Mulai Simulasi",
    "ats.loading": "Menganalisis kecocokan CV dengan Pekerjaan…",
    "ats.result.score": "Skor Kecocokan",
    "ats.result.missing": "Kata Kunci yang Hilang",
    "ats.result.recommendations": "Rekomendasi",
    "ats.history.title": "Riwayat Simulasi",
  },
};

// Inject toggle styles once
(function injectI18nStyles() {
  if (document.getElementById("i18n-styles")) return;
  const style = document.createElement("style");
  style.id = "i18n-styles";
  style.textContent = `
    .lang-switch {
      display: inline-flex;
      align-items: center;
      gap: 0;
      background: var(--bg, #f5f0e8);
      border: 3px solid var(--border, #1a1a1a);
      border-radius: 0;
      padding: 0;
      box-shadow: 2px 2px 0 var(--border, #1a1a1a);
      overflow: hidden;
    }
    .lang-switch .lang-btn {
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 6px 12px;
      font-family: var(--font-mono, 'Fira Code', monospace);
      font-size: 12px;
      font-weight: 700;
      text-transform: uppercase;
      cursor: pointer;
      border: none;
      background: transparent;
      color: var(--text, #1a1a1a);
      transition: background 0.15s ease, color 0.15s ease;
      letter-spacing: 0.03em;
      line-height: 1;
    }
    .lang-switch .lang-btn:not(:last-child) {
      border-right: 3px solid var(--border, #1a1a1a);
    }
    .lang-switch .lang-btn:hover {
      background: var(--accent4, #f0e6d3);
    }
    .lang-switch .lang-btn.active {
      background: var(--accent, #e74d3c);
      color: #fff;
    }
    .lang-switch .lang-btn .flag {
      font-size: 16px;
      line-height: 1;
    }
  `;
  document.head.appendChild(style);
})();

function setLanguage(lang) {
  if (!dictionary[lang]) return;

  localStorage.setItem("atstex_lang", lang);

  document.querySelectorAll("[data-i18n]").forEach((el) => {
    const key = el.getAttribute("data-i18n");
    if (dictionary[lang][key]) {
      el.textContent = dictionary[lang][key];
    }
  });

  // Special hero title handling (spans inside h1)
  const hero1 = document.querySelector('[data-i18n-hero="1"]');
  const hero2 = document.querySelector('[data-i18n-hero="2"]');
  const hero3 = document.querySelector('[data-i18n-hero="3"]');
  if (hero1 && hero2 && hero3) {
    hero1.textContent = dictionary[lang]["hero.title.1"];
    hero2.textContent = dictionary[lang]["hero.title.2"];
    hero3.textContent = dictionary[lang]["hero.title.3"];
  }

  // Update active toggle styling
  document
    .querySelectorAll(".lang-btn")
    .forEach((btn) => btn.classList.remove("active"));
  document
    .querySelectorAll(`[data-lang="${lang}"]`)
    .forEach((btn) => btn.classList.add("active"));
}

function initI18n() {
  const savedLang = localStorage.getItem("atstex_lang") || "id";

  // Bind click handlers
  document.querySelectorAll(".lang-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      const lang = btn.getAttribute("data-lang");
      if (lang) setLanguage(lang);
    });
  });

  setLanguage(savedLang);
}

document.addEventListener("DOMContentLoaded", initI18n);
