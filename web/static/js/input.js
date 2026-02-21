// ── Constants & Helpers ───────────────────────────────
const STORAGE_KEY = 'cv_data';
const ACTIVE_PROFILE_KEY = 'cv_active_profile_id';

// Default initial state
let data = {
  personal: {
    linkedin: {},
    github: {},
    website: {}
  },
  summary: "",
  experience: [],
  education: [],
  projects: [],
  skills: {},
  certifications: [],
  volunteer: [],
  awards: [],
  talks: []
};

let activeProfileId = null;

// ── Item Templates ──────────────────────────────────────
const itemTemplates = {
  experience: `
    <div class="dynamic-item-header">
      <span>Experience</span>
      <button class="btn btn-danger btn-remove" tabindex="-1">✕ Remove</button>
    </div>
    <div class="grid-2">
      <div class="form-group"><label>Company/Org</label><input type="text" data-dyn="company" placeholder="Google"></div>
      <div class="form-group"><label>Job Title</label><input type="text" data-dyn="title" placeholder="Senior Engineer"></div>
      <div class="form-group"><label>Location</label><input type="text" data-dyn="location" placeholder="New York, NY"></div>
      <div class="form-group"><label>Dates</label><input type="text" data-dyn="dates" placeholder="Jan 2020 ˝˝Present"></div>
    </div>
    <div class="form-group">
      <label>Bullet Points (one per line)</label>
      <textarea data-dyn="bullets" rows="10" placeholder="- Developed highly scalable microservices..."></textarea>
    </div>
  `,
  education: `
    <div class="dynamic-item-header">
      <span>Education</span>
      <button class="btn btn-danger btn-remove" tabindex="-1">✕ Remove</button>
    </div>
    <div class="grid-2">
      <div class="form-group"><label>Institution</label><input type="text" data-dyn="institution" placeholder="University of ..."></div>
      <div class="form-group"><label>Degree / Field</label><input type="text" data-dyn="degree" placeholder="B.S. Computer Science"></div>
      <div class="form-group"><label>Location</label><input type="text" data-dyn="location" placeholder="City, State"></div>
      <div class="form-group"><label>Dates</label><input type="text" data-dyn="dates" placeholder="Aug 2018 - May 2022"></div>
      <div class="form-group"><label>GPA / Honors (Optional)</label><input type="text" data-dyn="gpa" placeholder="3.8/4.0"></div>
    </div>
    <div class="form-group">
      <label>Activities / Involvement (Optional)</label>
      <input type="text" data-dyn="activities" placeholder="Student Union, Research Group...">
    </div>
  `,
  projects: `
    <div class="dynamic-item-header">
      <span>Project</span>
      <button class="btn btn-danger btn-remove" tabindex="-1">✕ Remove</button>
    </div>
    <div class="grid-2">
      <div class="form-group"><label>Project Name</label><input type="text" data-dyn="name" placeholder="Open Source DBMS"></div>
      <div class="form-group"><label>Role / Tech Stack</label><input type="text" data-dyn="role" placeholder="Creator / C++, Go"></div>
      <div class="form-group"><label>Link (Optional)</label><input type="text" data-dyn="link" placeholder="github.com/.../dbms"></div>
    </div>
    <div class="form-group">
      <label>Bullet Points (one per line)</label>
      <textarea data-dyn="bullets" rows="10" placeholder="- Engineered a specialized storage engine..."></textarea>
    </div>
  `,
  certifications: `
    <div class="dynamic-item-header">
      <span>Certification</span>
      <button class="btn btn-danger btn-remove" tabindex="-1">✕ Remove</button>
    </div>
    <div class="grid-2">
      <div class="form-group"><label>Name</label><input type="text" data-dyn="name" placeholder="AWS Certified Solutions Architect"></div>
      <div class="form-group"><label>Issuer & Date</label><input type="text" data-dyn="issuer" placeholder="Amazon, 2023"></div>
    </div>
  `,
  volunteer: `
    <div class="dynamic-item-header">
      <span>Volunteer</span>
      <button class="btn btn-danger btn-remove" tabindex="-1">✕ Remove</button>
    </div>
    <div class="grid-2">
      <div class="form-group"><label>Organization</label><input type="text" data-dyn="organization" placeholder="Red Cross"></div>
      <div class="form-group"><label>Role</label><input type="text" data-dyn="role" placeholder="Volunteer Field Worker"></div>
      <div class="form-group"><label>Location</label><input type="text" data-dyn="location" placeholder="Local City"></div>
      <div class="form-group"><label>Dates</label><input type="text" data-dyn="dates" placeholder="2019 - 2020"></div>
    </div>
    <div class="form-group">
      <label>Bullet Points (one per line)</label>
      <textarea data-dyn="bullets" rows="10" placeholder="- Assisted local community..."></textarea>
    </div>
  `,
  awards: `
    <div class="dynamic-item-header">
      <span>Award/Honor</span>
      <button class="btn btn-danger btn-remove" tabindex="-1">✕ Remove</button>
    </div>
    <div class="grid-2">
      <div class="form-group"><label>Title</label><input type="text" data-dyn="title" placeholder="Employee of the Year"></div>
      <div class="form-group"><label>Issuer</label><input type="text" data-dyn="issuer" placeholder="Company Name"></div>
      <div class="form-group"><label>Date</label><input type="text" data-dyn="date" placeholder="Dec 2021"></div>
    </div>
    <div class="form-group">
      <label>Description (Optional)</label>
      <input type="text" data-dyn="description" placeholder="Awarded for exceeding target goals...">
    </div>
  `,
  talks: `
    <div class="dynamic-item-header">
      <span>Talk/Conference</span>
      <button class="btn btn-danger btn-remove" tabindex="-1">✕ Remove</button>
    </div>
    <div class="grid-2">
      <div class="form-group"><label>Title</label><input type="text" data-dyn="title" placeholder="Scaling DBs"></div>
      <div class="form-group"><label>Event</label><input type="text" data-dyn="event" placeholder="TechConf 2023"></div>
      <div class="form-group"><label>Location</label><input type="text" data-dyn="location" placeholder="San Francisco"></div>
      <div class="form-group"><label>Date</label><input type="text" data-dyn="date" placeholder="Oct 2023"></div>
    </div>
    <div class="form-group">
      <label>Description/Link (Optional)</label>
      <input type="text" data-dyn="description" placeholder="Presented to 500+ attendees...">
    </div>
  `
};

// ── Serialization & Deserialization ────────────────────
function collectFormData() {
  // Capture static fields
  document.querySelectorAll('input[data-field], textarea[data-field]').forEach(el => {
    const parts = el.dataset.field.split('.');
    let current = data;
    for (let i = 0; i < parts.length - 1; i++) {
        if (!current[parts[i]]) {
            current[parts[i]] = {};
        }
        current = current[parts[i]];
    }
    current[parts[parts.length - 1]] = el.value;
  });

  // Capture dynamic lists
  ['experience', 'education', 'projects', 'certifications', 'volunteer', 'awards', 'talks'].forEach(key => {
    data[key] = [];
    const container = document.getElementById(`list-${key}`);
    if (!container) return;
    container.querySelectorAll('.dynamic-item').forEach(itemEl => {
      const obj = {};
      itemEl.querySelectorAll('[data-dyn]').forEach(input => {
        obj[input.dataset.dyn] = input.value;
      });
      data[key].push(obj);
    });
  });

  return data;
}

function saveToStorage() {
  collectFormData();
  localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
}

function clearForm() {
  // Reset data model
  data = {
    personal: { linkedin: {}, github: {}, website: {} },
    summary: "",
    experience: [], education: [], projects: [],
    skills: {},
    certifications: [], volunteer: [], awards: [], talks: []
  };

  // Clear static fields
  document.querySelectorAll('input[data-field], textarea[data-field]').forEach(el => {
    el.value = '';
  });

  // Clear dynamic lists
  ['experience', 'education', 'projects', 'certifications', 'volunteer', 'awards', 'talks'].forEach(key => {
    const container = document.getElementById(`list-${key}`);
    if (container) container.innerHTML = '';
  });
}

function populateForm(newData) {
  clearForm();
  if (!newData) return;
  data = newData;

  // Populate static fields
  document.querySelectorAll('input[data-field], textarea[data-field]').forEach(el => {
    const parts = el.dataset.field.split('.');
    let current = data;
    let found = true;
    for (let i = 0; i < parts.length; i++) {
        if (current[parts[i]] === undefined) {
            found = false;
            break;
        }
        current = current[parts[i]];
    }
    if (found) {
        el.value = current;
    }
  });

  // Populate dynamic lists
  ['experience', 'education', 'projects', 'certifications', 'volunteer', 'awards', 'talks'].forEach(key => {
    const items = data[key] || [];
    items.forEach(itemData => addItem(key, itemData));
  });
}

function loadFromStorage() {
  const jsonStr = localStorage.getItem(STORAGE_KEY);
  if (!jsonStr) return;
  try {
    const parsed = JSON.parse(jsonStr);
    populateForm(parsed);
  } catch(e) { console.error(e); }
}

// ── DOM Interactions ───────────────────────────────────
function addItem(type, initialData = null) {
  const container = document.getElementById(`list-${type}`);
  if (!container || !itemTemplates[type]) return;

  const div = document.createElement('div');
  div.className = 'dynamic-item';
  div.innerHTML = itemTemplates[type];

  // Attach delete behavior
  div.querySelector('.btn-remove').addEventListener('click', () => {
    div.remove();
    saveToStorage();
  });

  // Attach change listeners to new inputs
  div.querySelectorAll('input, textarea').forEach(el => {
    el.addEventListener('input', saveToStorage);
  });

  // Pre-fill if load
  if (initialData) {
    div.querySelectorAll('[data-dyn]').forEach(el => {
      const key = el.dataset.dyn;
      if (initialData[key] !== undefined) el.value = initialData[key];
    });
  }

  container.appendChild(div);

  if (!initialData) { // focus first input if user added manually
    const firstInput = div.querySelector('input');
    if (firstInput) firstInput.focus();
  }
}

// Attach listeners to Add buttons
document.querySelectorAll('.btn-add').forEach(btn => {
  btn.addEventListener('click', () => {
    addItem(btn.dataset.target);
    saveToStorage();
  });
});

// Attach general listeners to static fields
document.querySelectorAll('input[data-field], textarea[data-field]').forEach(el => {
  el.addEventListener('input', saveToStorage);
});

// ── CV Profile Management ──────────────────────────────
const profileSelect = document.getElementById('cv-profile-select');
const btnNewProfile = document.getElementById('btn-new-profile');
const btnSaveDB = document.getElementById('btn-save-db');
const btnDeleteProfile = document.getElementById('btn-delete-profile');
const btnFillDummy = document.getElementById('btn-fill-dummy');
const btnUploadPdf = document.getElementById('btn-upload-pdf');
const pdfUploadInput = document.getElementById('pdf-upload');
const statusMsg = document.getElementById('cv-status-msg');

// ── Dummy Data ─────────────────────────────────────────
const DUMMY_BIODATA = {
  personal: {
    name: "Sammi Aldhi Yanto",
    title: "Senior Software Engineer",
    email: "sammi.aldhi@gmail.com",
    phone: "+62 812 3456 7890",
    location: "Jakarta, Indonesia",
    linkedin: {
      display: "linkedin.com/in/sammidev",
      url: "https://linkedin.com/in/sammidev"
    },
    github: {
      display: "github.com/SemmiDev",
      url: "https://github.com/SemmiDev"
    },
    website: {
      display: "sammidev.com",
      url: "https://sammidev.com"
    }
  },
  summary: "Senior Software Engineer with 8+ years of experience building large-scale distributed systems at top technology companies. Proven track record of designing and shipping products used by billions of users worldwide. Deep expertise in backend engineering, cloud infrastructure, and system architecture. Passionate about open-source contributions and mentoring the next generation of engineers.",
  experience: [
    {
      company: "Google",
      title: "Senior Software Engineer — Cloud Infrastructure",
      location: "Mountain View, CA (Remote)",
      dates: "Jan 2022 — Present",
      bullets: "Architected and led the migration of a monolithic API gateway serving 2B+ daily requests to a microservices-based architecture using Go and gRPC, reducing p99 latency by 40%\nDesigned a distributed caching layer with Redis Cluster and Memcached that reduced database load by 65% across 12 services\nBuilt a real-time anomaly detection pipeline using Apache Beam and BigQuery, processing 500K events/sec to identify SLA violations\nMentored 5 junior engineers through Google's Engineering Residency program and led weekly architecture review sessions\nContributed to internal Go standard library improvements adopted across 200+ teams"
    },
    {
      company: "Meta (Facebook)",
      title: "Software Engineer — News Feed Ranking",
      location: "Menlo Park, CA",
      dates: "Jun 2019 — Dec 2021",
      bullets: "Developed and optimized the News Feed content ranking pipeline serving 3.5B monthly active users using C++ and Python\nImplemented an A/B testing framework that reduced experiment setup time from 2 weeks to 2 hours, enabling 300+ concurrent experiments\nBuilt a feature store using Apache Spark and Hive to serve ML models with sub-10ms latency at scale\nCollaborated with the Integrity team to develop automated content moderation systems reducing harmful content by 35%\nReceived the 'Impact Award' for shipping a personalization algorithm that increased user engagement by 12%"
    },
    {
      company: "Amazon Web Services (AWS)",
      title: "Software Development Engineer II — DynamoDB",
      location: "Seattle, WA",
      dates: "Aug 2017 — May 2019",
      bullets: "Contributed to DynamoDB's core storage engine, implementing adaptive capacity allocation that improved throughput for bursty workloads by 50%\nDesigned and built an automated partition management system that handles 10+ trillion API calls per day\nDeveloped chaos engineering tools used by 50+ internal teams to validate service resilience\nLed the migration of monitoring infrastructure from legacy systems to CloudWatch, reducing operational overhead by 30%\nAuthored 3 internal technical papers on distributed consensus protocols adopted as reference material"
    },
    {
      company: "Apple",
      title: "Software Engineer — Siri Backend",
      location: "Cupertino, CA",
      dates: "Jul 2016 — Jul 2017",
      bullets: "Built low-latency natural language processing microservices in Java and Swift handling 500M+ daily Siri queries\nOptimized the intent classification pipeline, reducing inference time by 25% while maintaining 98.5% accuracy\nDeveloped automated integration testing framework that cut release cycle regression testing from 3 days to 4 hours\nCollaborated with the ML team to deploy on-device models that reduced server-side query volume by 20%"
    }
  ],
  education: [
    {
      institution: "Stanford University",
      degree: "M.S. Computer Science — Distributed Systems",
      location: "Stanford, CA",
      dates: "Sep 2014 — Jun 2016",
      gpa: "3.92 / 4.0",
      activities: "Teaching Assistant for CS244b (Distributed Systems), Stanford ACM Chapter VP"
    },
    {
      institution: "Institut Teknologi Bandung (ITB)",
      degree: "B.S. Informatics Engineering",
      location: "Bandung, Indonesia",
      dates: "Aug 2010 — Jun 2014",
      gpa: "3.85 / 4.0 — Cum Laude",
      activities: "Competitive Programming Team Captain, Google Developer Student Club Lead"
    }
  ],
  projects: [
    {
      name: "AtstexLab",
      role: "Creator — Go, LaTeX, Tailwind CSS",
      link: "github.com/SemmiDev/atstex-lab",
      bullets: "Built a self-hosted ATS-friendly resume builder with live LaTeX compilation and PDF preview\nImplemented multi-CV profile management with PostgreSQL and Google OAuth integration\nDesigned a brutalist UI with responsive sidebar, tabbed editor, and embedded biodata form"
    },
    {
      name: "DistKV",
      role: "Creator — Go, Raft Consensus",
      link: "github.com/SemmiDev/distkv",
      bullets: "Engineered a distributed key-value store with Raft consensus supporting 100K+ ops/sec\nImplemented log compaction, snapshotting, and membership change protocols from scratch\nAchieved 99.99% availability in production workloads across 5-node clusters"
    },
    {
      name: "GoQueue",
      role: "Creator — Go, Protocol Buffers",
      link: "github.com/SemmiDev/goqueue",
      bullets: "Built a high-performance message queue with exactly-once delivery semantics\nSupports persistent storage with write-ahead log and configurable retention policies\nBenchmarked at 250K messages/sec with sub-millisecond p99 publish latency"
    }
  ],
  skills: {
    languages: "Go, Python, Java, C++, TypeScript, Rust, SQL",
    frameworks: "gRPC, Gin, React, Next.js, Apache Beam, Apache Spark, Kubernetes",
    tools: "Docker, Terraform, AWS (DynamoDB, Lambda, S3, ECS), GCP (BigQuery, Pub/Sub, GKE), PostgreSQL, Redis, Kafka",
    other: "Distributed Systems Design, System Architecture, Technical Leadership, Mentoring, Agile, CI/CD"
  },
  certifications: [
    {
      name: "AWS Certified Solutions Architect — Professional",
      issuer: "Amazon Web Services, 2023"
    },
    {
      name: "Google Cloud Professional Cloud Architect",
      issuer: "Google Cloud, 2022"
    },
    {
      name: "Certified Kubernetes Administrator (CKA)",
      issuer: "Cloud Native Computing Foundation, 2021"
    }
  ],
  volunteer: [
    {
      organization: "Code.org Indonesia",
      role: "Lead Instructor",
      location: "Jakarta, Indonesia",
      dates: "2020 — Present",
      bullets: "Taught programming fundamentals to 500+ underprivileged high school students across 15 schools\nDeveloped a Go programming curriculum adapted for Bahasa Indonesia\nOrganized annual hackathons connecting students with industry mentors"
    },
    {
      organization: "Google Summer of Code",
      role: "Mentor — Go Project",
      location: "Remote",
      dates: "2021 — 2023",
      bullets: "Mentored 6 open-source contributors on the Go compiler and standard library\nGuided students through code review, testing best practices, and community engagement\n4 of 6 mentees became regular Go contributors post-program"
    }
  ],
  awards: [
    {
      title: "Meta Impact Award",
      issuer: "Meta Platforms",
      date: "Dec 2021",
      description: "Awarded for shipping a personalization algorithm that increased News Feed engagement by 12% across 3.5B users"
    },
    {
      title: "AWS Builder Award",
      issuer: "Amazon Web Services",
      date: "Mar 2019",
      description: "Recognized for contributions to DynamoDB's adaptive capacity system handling 10+ trillion daily API calls"
    },
    {
      title: "ICPC Asia Regional — Gold Medal",
      issuer: "ACM International Collegiate Programming Contest",
      date: "Nov 2013",
      description: "1st place in the Indonesia National Contest and Gold Medal in the Asia Regional Finals representing ITB"
    }
  ],
  talks: [
    {
      title: "Building Resilient Distributed Systems in Go",
      event: "GopherCon 2023",
      location: "San Diego, CA",
      date: "Sep 2023",
      description: "Keynote talk on designing fault-tolerant microservices using Go, presented to 2,000+ attendees. Covered circuit breakers, bulkheads, and chaos engineering patterns."
    },
    {
      title: "Scaling DynamoDB to 10 Trillion Requests",
      event: "AWS re:Invent 2018",
      location: "Las Vegas, NV",
      date: "Nov 2018",
      description: "Technical deep-dive on DynamoDB's partition management and adaptive capacity allocation, co-presented with the DynamoDB principal engineer."
    },
    {
      title: "Open Source in Southeast Asia: Building Communities",
      event: "FOSSASIA Summit 2022",
      location: "Singapore",
      date: "Apr 2022",
      description: "Panel discussion on growing open-source communities in emerging tech ecosystems, moderated a workshop on first contributions."
    }
  ]
};

function showStatus(msg, isError = false) {
  statusMsg.textContent = msg;
  statusMsg.classList.remove('hidden', '!text-error', '!text-accent2');
  statusMsg.classList.add(isError ? '!text-error' : '!text-accent2');
  setTimeout(() => statusMsg.classList.add('hidden'), 3000);
}

function updateButtons() {
  const hasProfile = !!activeProfileId;
  btnSaveDB.disabled = !hasProfile;
  btnDeleteProfile.disabled = !hasProfile;
  if (btnFillDummy) btnFillDummy.disabled = !hasProfile;
  if (btnUploadPdf) btnUploadPdf.disabled = !hasProfile;
}

async function fetchProfiles() {
  try {
    const res = await fetch('/api/cv-profiles');
    if (!res.ok) return [];
    return await res.json();
  } catch (e) {
    console.error('Failed to fetch profiles', e);
    return [];
  }
}

async function loadProfileList() {
  const profiles = await fetchProfiles();

  // Rebuild select options
  profileSelect.innerHTML = '<option value="">— Select or Create —</option>';
  profiles.forEach(p => {
    const opt = document.createElement('option');
    opt.value = p.id;
    opt.textContent = p.title;
    profileSelect.appendChild(opt);
  });

  // Restore previously active profile
  const savedId = localStorage.getItem(ACTIVE_PROFILE_KEY);
  if (savedId && profiles.find(p => p.id === savedId)) {
    profileSelect.value = savedId;
    activeProfileId = savedId;
    await loadProfileData(savedId);
  } else {
    // No saved profile — just load from localStorage
    loadFromStorage();
  }

  updateButtons();
}

async function loadProfileData(profileId) {
  try {
    const res = await fetch(`/api/cv-profiles/${profileId}`);
    if (!res.ok) throw new Error('Failed to load profile');
    const profile = await res.json();

    // Parse biodata and populate form
    let biodata = profile.biodata;
    if (typeof biodata === 'string') {
      biodata = JSON.parse(biodata);
    }

    // Merge with default structure
    const fullData = {
      personal: { linkedin: {}, github: {}, website: {}, ...((biodata && biodata.personal) || {}) },
      summary: (biodata && biodata.summary) || "",
      experience: (biodata && biodata.experience) || [],
      education: (biodata && biodata.education) || [],
      projects: (biodata && biodata.projects) || [],
      skills: (biodata && biodata.skills) || {},
      certifications: (biodata && biodata.certifications) || [],
      volunteer: (biodata && biodata.volunteer) || [],
      awards: (biodata && biodata.awards) || [],
      talks: (biodata && biodata.talks) || []
    };

    populateForm(fullData);
    saveToStorage(); // sync to localStorage
    showStatus(`Loaded "${profile.title}" from database`);
  } catch (e) {
    console.error(e);
    showStatus('Failed to load profile', true);
  }
}

// On dropdown change
profileSelect.addEventListener('change', async () => {
  const id = profileSelect.value;
  if (!id) {
    activeProfileId = null;
    localStorage.removeItem(ACTIVE_PROFILE_KEY);
    clearForm();
    updateButtons();
    return;
  }

  activeProfileId = id;
  localStorage.setItem(ACTIVE_PROFILE_KEY, id);
  await loadProfileData(id);
  updateButtons();
});

// New Profile
btnNewProfile.addEventListener('click', async () => {
  const title = prompt('Enter a title for this CV profile (e.g., "Back End Developer"):');
  if (!title || !title.trim()) return;

  try {
    const res = await fetch('/api/cv-profiles', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: title.trim() })
    });

    if (!res.ok) throw new Error('Failed to create profile');
    const profile = await res.json();

    // Add to dropdown and select it
    const opt = document.createElement('option');
    opt.value = profile.id;
    opt.textContent = profile.title;
    profileSelect.appendChild(opt);
    profileSelect.value = profile.id;

    activeProfileId = profile.id;
    localStorage.setItem(ACTIVE_PROFILE_KEY, profile.id);

    // Clear form for new profile
    clearForm();
    saveToStorage();
    updateButtons();

    showStatus(`Created "${profile.title}" — start filling in your data!`);
  } catch (e) {
    console.error(e);
    showStatus('Failed to create profile', true);
  }
});

// Save to DB
btnSaveDB.addEventListener('click', async () => {
  if (!activeProfileId) return;

  collectFormData();

  try {
    const res = await fetch(`/api/cv-profiles/${activeProfileId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ biodata: data })
    });

    if (!res.ok) throw new Error('Failed to save');
    showStatus('✅ Saved to database successfully!');
  } catch (e) {
    console.error(e);
    showStatus('Failed to save to database', true);
  }
});

// Delete Profile
btnDeleteProfile.addEventListener('click', async () => {
  if (!activeProfileId) return;

  const selectedOpt = profileSelect.options[profileSelect.selectedIndex];
  const title = selectedOpt ? selectedOpt.textContent : 'this profile';

  if (!confirm(`Delete "${title}"? This cannot be undone.`)) return;

  try {
    const res = await fetch(`/api/cv-profiles/${activeProfileId}`, {
      method: 'DELETE'
    });

    if (!res.ok) throw new Error('Failed to delete');

    // Remove from dropdown
    if (selectedOpt) selectedOpt.remove();

    activeProfileId = null;
    localStorage.removeItem(ACTIVE_PROFILE_KEY);
    profileSelect.value = '';
    clearForm();
    saveToStorage();
    updateButtons();

    showStatus(`Deleted "${title}"`);
  } catch (e) {
    console.error(e);
    showStatus('Failed to delete profile', true);
  }
});

// Fill Dummy Data
if (btnFillDummy) {
  btnFillDummy.addEventListener('click', () => {
    if (!activeProfileId) {
      showStatus('Select or create a profile first', true);
      return;
    }
    populateForm(JSON.parse(JSON.stringify(DUMMY_BIODATA)));
    saveToStorage();
    showStatus('Filled with example data — Sammi Aldhi Yanto 🚀');
  });
}

// ── PDF Extraction ─────────────────────────────────────
if (btnUploadPdf && pdfUploadInput) {
  btnUploadPdf.addEventListener('click', () => {
    if (!activeProfileId) {
      showStatus('Select or create a profile first', true);
      return;
    }
    pdfUploadInput.click();
  });

  pdfUploadInput.addEventListener('change', async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    if (file.type !== 'application/pdf') {
      showStatus('Please upload a valid PDF file', true);
      return;
    }

    const originalBtnHTML = btnUploadPdf.innerHTML;
    btnUploadPdf.innerHTML = '<i class="ph-bold ph-spinner animate-spin"></i> Extracting...';
    btnUploadPdf.disabled = true;

    try {
      showStatus('Reading PDF... 📄');
      // 1. Read PDF with PDF.js
      const arrayBuffer = await file.arrayBuffer();
      const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;
      let fullText = '';

      for (let i = 1; i <= pdf.numPages; i++) {
        const page = await pdf.getPage(i);
        const textContent = await page.getTextContent();
        const pageText = textContent.items.map(item => item.str).join(' ');
        fullText += pageText + '\n';
      }

      const textLen = fullText.trim().length;
      if (textLen < 50) {
        throw new Error('Could not extract enough text from the PDF');
      }

      showStatus(`Processing ${textLen} characters with AI... 🤖`);

      // 2. Send to backend AI extractor
      const res = await fetch('/api/extract-pdf', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: fullText })
      });

      if (!res.ok) {
        const errData = await res.json();
        throw new Error(errData.error || 'Failed to extract data via API');
      }

      const extractedData = await res.json();

      // 3. Populate form and save
      populateForm(extractedData);
      saveToStorage();
      showStatus('✨ Successfully extracted and applied data from PDF!', false);

    } catch (err) {
      console.error('PDF extraction error:', err);
      showStatus(err.message || 'Error occurred during PDF extraction', true);
    } finally {
      // Reset button
      btnUploadPdf.innerHTML = originalBtnHTML;
      btnUploadPdf.disabled = false;
      pdfUploadInput.value = ''; // clear input so the same file can be selected again
    }
  });
}

// ── Init ───────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  loadProfileList();
});
