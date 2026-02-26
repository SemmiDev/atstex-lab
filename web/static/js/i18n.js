// i18n.js - Lightweight client-side localization

const dictionary = {
  en: {
    "nav.home": "Home",
    "nav.build": "Build",
    "hero.badge": "Resume Builder That Beats the Bots",
    "hero.title.1": "ATS",
    "hero.title.2": "Optimized",
    "hero.title.3": "& Recruiter-Centric.",
    "hero.subtitle": "ATSTEXLAB helps you build professional, ATS-friendly resumes that actually get read by recruiters. Fill in your details, pick a template, and download a polished PDF.",
    "hero.btn.builder": "Builder",
    "hero.btn.logout": "Logout",
    "hero.btn.signin": "Sign in with Google",
    "hero.trusted": "TRUSTED BY JOB SEEKERS, FRESH GRADS, AND CAREER SWITCHERS",
    "feature.1.title": "Instant PDF Export",
    "feature.1.desc": "Hit compile and get a perfectly formatted PDF in seconds. No software to install — everything runs in your browser.",
    "feature.2.title": "100% Free, Forever",
    "feature.2.desc": "No paywalls, no subscriptions, no hidden fees. Sign in with Google to save your CV profiles — completely free.",
    "feature.3.title": "ATS-Friendly Templates",
    "feature.3.desc": "Every template is designed to pass Applicant Tracking Systems. Clean layout, clear sections, no hidden formatting tricks.",
    "feature.4.title": "Multiple CV Profiles",
    "feature.4.desc": "Create different resumes for different roles. One for \"Back End Developer\", another for \"Data Engineer\" — you decide.",
    "feature.5.title": "AI-Powered PDF Extraction",
    "feature.5.desc": "Upload your existing PDF resume and let our local AI parse it directly into your biodata profile. No manual copy-pasting required.",
    "feature.6.title": "AI CV Review & Scoring",
    "feature.6.desc": "Get actionable feedback tailored by AI. Discover strengths, weaknesses, and a structured rating to help your resume stand out.",
    "feature.7.title": "AI Cover Letter Generator",
    "feature.7.desc": "Generate highly tailored cover letters combining your structured CV data and a specific job description.",
    "support.title": "Support the Project",
    "support.desc": "If you found this tool helpful, consider leaving a tip to keep the servers running!",
    "footer.copyright": "© 2026 ATSTEXLAB · Your Resume, Recruiter-Ready"
  },
  id: {
    "nav.home": "Beranda",
    "nav.build": "Buat CV",
    "hero.badge": "Pembuat CV yang Lolos Screening Robot",
    "hero.title.1": "Lolos Seleksi",
    "hero.title.2": "Sistem.",
    "hero.title.3": "Memikat Hati Rekruter.",
    "hero.subtitle": "ATSTEXLAB membantu Anda membuat CV profesional yang ramah ATS dan disukai rekruter. Isi data, pilih template, dan unduh PDF yang rapi.",
    "hero.btn.builder": "Buat CV",
    "hero.btn.logout": "Keluar",
    "hero.btn.signin": "Masuk dengan Google",
    "hero.trusted": "DIPERCAYA OLEH PENCARI KERJA, LULUSAN BARU, DAN PROFESIONAL",
    "feature.1.title": "Ekspor PDF Instan",
    "feature.1.desc": "Klik compile dan dapatkan PDF berformat sempurna dalam hitungan detik. Tanpa instalasi software — semuanya berjalan di browser.",
    "feature.2.title": "100% Gratis, Selamanya",
    "feature.2.desc": "Tanpa biaya langganan, tanpa fitur tersembunyi. Masuk dengan Google untuk menyimpan profil CV Anda — sepenuhnya gratis.",
    "feature.3.title": "Template Ramah ATS",
    "feature.3.desc": "Setiap template didesain untuk lolos Applicant Tracking Systems. Tata letak bersih, bagian yang jelas, tanpa trik format tersembunyi.",
    "feature.4.title": "Banyak Profil CV",
    "feature.4.desc": "Buat CV berbeda untuk peran yang berbeda. Satu untuk \"Back End Developer\", satu lagi untuk \"Data Engineer\" — Anda yang tentukan.",
    "feature.5.title": "Ekstraksi PDF dengan AI",
    "feature.5.desc": "Unggah CV PDF Anda yang sudah ada dan biarkan AI kami menyalinnya langsung ke profil biodata Anda. Bebas repot salin-tempel manual.",
    "feature.6.title": "Review & Penilaian CV dengan AI",
    "feature.6.desc": "Dapatkan masukan dari AI. Ketahui kelebihan, kekurangan, dan skor terstruktur agar CV Anda lebih menonjol.",
    "feature.7.title": "Pembuat Cover Letter AI",
    "feature.7.desc": "Buat cover letter yang sangat sesuai dengan menggabungkan data CV Anda dan deskripsi pekerjaan secara spesifik.",
    "support.title": "Dukung Proyek Ini",
    "support.desc": "Jika alat ini bermanfaat bagi Anda, pertimbangkan untuk memberikan tip agar server tetap menyala!",
    "footer.copyright": "© 2026 ATSTEXLAB · CV Anda, Siap Dipanggil"
  }
};

// Inject toggle styles once
(function injectI18nStyles() {
  if (document.getElementById('i18n-styles')) return;
  const style = document.createElement('style');
  style.id = 'i18n-styles';
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

  document.querySelectorAll("[data-i18n]").forEach(el => {
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
  document.querySelectorAll('.lang-btn').forEach(btn => btn.classList.remove('active'));
  document.querySelectorAll(`[data-lang="${lang}"]`).forEach(btn => btn.classList.add('active'));
}

function initI18n() {
  const savedLang = localStorage.getItem("atstex_lang") || "id";

  // Bind click handlers
  document.querySelectorAll('.lang-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const lang = btn.getAttribute('data-lang');
      if (lang) setLanguage(lang);
    });
  });

  setLanguage(savedLang);
}

document.addEventListener("DOMContentLoaded", initI18n);
