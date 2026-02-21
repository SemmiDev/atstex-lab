// i18n.js - Lightweight client-side localization

const dictionary = {
  en: {
    "nav.home": "Home",
    "nav.build": "Build",
    "hero.badge": "Resume Builder That Beats the Bots",
    "hero.title.1": "Structured for",
    "hero.title.2": "Machines.",
    "hero.title.3": "Designed for Humans.",
    "hero.subtitle": "ATSTEXLAB helps you build professional, ATS-friendly resumes that actually get read by recruiters. Fill in your details, pick a template, and download a polished PDF — all from your browser.",
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
    "support.title": "Support the Project",
    "support.desc": "If you found this tool helpful, consider leaving a tip to keep the servers running!",
    "footer.copyright": "© 2026 ATSTEXLAB · Your Resume, Recruiter-Ready"
  },
  id: {
    "nav.home": "Beranda",
    "nav.build": "Buat CV",
    "hero.badge": "Pembuat CV yang Lolos Screening Robot",
    "hero.title.1": "Terstruktur untuk",
    "hero.title.2": "Mesin.",
    "hero.title.3": "Didesain untuk Manusia.",
    "hero.subtitle": "ATSTEXLAB membantu Anda membuat CV profesional yang ramah ATS dan disukai rekruter. Isi data, pilih template, dan unduh PDF yang rapi — semua dari browser Anda.",
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
    "support.title": "Dukung Proyek Ini",
    "support.desc": "Jika alat ini bermanfaat bagi Anda, pertimbangkan untuk memberikan tip agar server tetap menyala!",
    "footer.copyright": "© 2026 ATSTEXLAB · CV Anda, Siap Dipanggil"
  }
};

function setLanguage(lang) {
  if (!dictionary[lang]) return;

  localStorage.setItem("atstex_lang", lang);

  document.querySelectorAll("[data-i18n]").forEach(el => {
    const key = el.getAttribute("data-i18n");
    if (dictionary[lang][key]) {
      // Small hack: if the text had HTML originally, innerHTML is fine.
      // But we mostly just have text. Check if it contains <br> or <span> logic.
      if (key.includes("hero.title")) {
          // handled specially since it has span
      } else {
        el.textContent = dictionary[lang][key];
      }
    }
  });

  // Special hero title handling
  const hero1 = document.querySelector('[data-i18n-hero="1"]');
  const hero2 = document.querySelector('[data-i18n-hero="2"]');
  const hero3 = document.querySelector('[data-i18n-hero="3"]');
  if (hero1 && hero2 && hero3) {
      hero1.textContent = dictionary[lang]["hero.title.1"];
      hero2.textContent = dictionary[lang]["hero.title.2"];
      hero3.textContent = dictionary[lang]["hero.title.3"];
  }

  // Update active toggle styling
  document.querySelectorAll('.lang-toggle').forEach(btn => btn.classList.remove('active-lang'));
  const activeBtn = document.getElementById(`lang-${lang}`);
  if (activeBtn) activeBtn.classList.add('active-lang');
}

function initI18n() {
  const savedLang = localStorage.getItem("atstex_lang") || "en";
  setLanguage(savedLang);

  const btnEn = document.getElementById("lang-en");
  const btnId = document.getElementById("lang-id");

  if (btnEn) btnEn.addEventListener("click", () => setLanguage("en"));
  if (btnId) btnId.addEventListener("click", () => setLanguage("id"));
}

document.addEventListener("DOMContentLoaded", initI18n);
