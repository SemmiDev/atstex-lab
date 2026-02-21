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
      <div class="form-group"><label>Dates</label><input type="text" data-dyn="dates" placeholder="Jan 2020 - Present"></div>
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
const statusMsg = document.getElementById('cv-status-msg');

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

// ── Init ───────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  loadProfileList();
});
