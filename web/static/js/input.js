// ── Constants & Helpers ───────────────────────────────
const STORAGE_KEY = 'cv_data';

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
      <textarea data-dyn="bullets" rows="4" placeholder="- Developed highly scalable microservices..."></textarea>
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
      <textarea data-dyn="bullets" rows="3" placeholder="- Engineered a specialized storage engine..."></textarea>
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
      <textarea data-dyn="bullets" rows="3" placeholder="- Assisted local community..."></textarea>
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
function saveToStorage() {
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

  localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
  console.log('Saved to standard localStorage:', data);
}

function loadFromStorage() {
  const jsonStr = localStorage.getItem(STORAGE_KEY);
  if (!jsonStr) return;
  try {
    data = JSON.parse(jsonStr);
  } catch(e) { console.error(e); return; }

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

// Attach listenets to Add buttons
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

// ── Init ───────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  loadFromStorage();
});
