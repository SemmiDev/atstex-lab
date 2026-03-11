# Promising Features for ATSTEX-LAB

Prioritized by **implementation effort** and **user impact**, grouped into tiers. Each feature notes what existing infrastructure it builds on.

> [!TIP]
> Features marked with ⚡ can be built **entirely in the frontend** (JS/CSS only, no backend changes). Features marked with 🤖 leverage your existing AI provider integration.

---

## 🟢 Tier 1 — Quick Wins (1-2 days each)

### 1. ⚡ Dark Mode Toggle
**Impact**: High (developer/power-user favorite)
Add a dark theme using CSS custom properties you already define in `:root`. Toggle persisted via `localStorage`. Your design system (`--bg`, `--surface`, `--accent`, etc.) makes this straightforward — just swap the values.

### 2. ⚡ CV Completeness Score Bar
**Impact**: Medium-High (gamification → more biodata filled → better CVs)
A visual progress bar on the `/input` page showing "Your CV is 65% complete". Score calculated client-side from filled JSONB fields (name, email, experience count, skills, etc.). Encourages users to fill optional sections.

### 3. ⚡ Template Preview Thumbnails in Editor Dropdown
**Impact**: Medium
Replace the plain `<select>` dropdown on the editor page with a visual picker showing small static preview images for each template (bay, delta, reef, sea, tide, wave). Uses the Gallery infrastructure you just built.

### 4. ⚡ Export Biodata as JSON / Import from JSON
**Impact**: Medium (portability, backup)
Add "Export" and "Import" buttons on `/input`. Export serializes `localStorage`'s `cv_data` to a [.json](file:///Users/sammidev/Developments/atstex-lab/package.json) file download. Import lets users drag-drop a JSON file to restore their data. Zero backend changes.

### 5. Compilation History Log
**Impact**: Medium
Store last N compilations (timestamp, template, elapsed, success/fail) in a PostgreSQL table. Show a small "History" panel in the editor. Builds on existing [compile](file:///Users/sammidev/Developments/atstex-lab/internal/handler/server.go#28-33) handler — just add a DB insert after successful compilation.

---

## 🟡 Tier 2 — Medium Effort (3-5 days each)

### 6. 🤖 AI Interview Prep
**Impact**: Very High (unique differentiator)
You already have the user's CV (`CVProfile.Biodata`) and job description (from `AtsSimulation.JobDescription`). Add a new page `/interview-prep` that sends both to your AI provider with a prompt to generate 8-10 tailored interview questions (behavioral + technical). New domain model `InterviewPrep`, new handler, new HTML page.

### 7. 🤖 AI Bullet Point Enhancer
**Impact**: High (directly improves CV quality)
On the `/input` page, add a small ✨ button next to each experience bullet textarea. Clicking it sends the raw bullet to AI with a prompt like "Rewrite this bullet point using strong action verbs and quantifiable metrics." Returns an improved version the user can accept or reject. Reuses your existing `extractor.AIConfig`.

### 8. PDF Watermark-Free Export (Freemium Gate)
**Impact**: High (monetization lever)
For free-tier users, add a subtle "Built with ATSTEXLAB" watermark to compiled PDFs. Pro users get clean exports. Implement by conditionally injecting a LaTeX footer line in `cvtemplate.Render()` based on subscription tier passed from the handler.

### 9. Application Deadline Tracking + Status Stats
**Impact**: Medium-High
Extend [JobApplication](file:///Users/sammidev/Developments/atstex-lab/internal/domain/job_application.go#9-20) with a `deadline` column (`DATE`). Show overdue applications in red on the Kanban board. Add a stats panel at the top: "12 Applied → 4 Interviews → 1 Offer" pipeline funnel. Builds on existing [job_application.go](file:///Users/sammidev/Developments/atstex-lab/internal/domain/job_application.go) and [kanban.js](file:///Users/sammidev/Developments/atstex-lab/web/static/js/kanban.js).

### 10. CV Version History / Snapshots
**Impact**: Medium-High (peace of mind)
Every time the user saves a CV profile, also store a snapshot in a `cv_profile_versions` table (profile_id, biodata JSONB, created_at). Add a "Version History" drawer on `/input` that lists snapshots and lets them restore any version. Prevents accidental data loss.

### 11. Template Color Customization
**Impact**: Medium (personalization)
Expose a "Primary Color" picker in the Page Settings panel. The chosen hex color gets passed to `cvtemplate.Render()` as part of [PageSettings](file:///Users/sammidev/Developments/atstex-lab/internal/cvtemplate/cvtemplate.go#145-159) and injected into the LaTeX template via `\definecolor`. Your templates already use `[[ ]]` Go template delimiters, so adding `[[ .Settings.PrimaryColor ]]` is simple.

---

## 🔴 Tier 3 — Larger Effort (1-2 weeks each)

### 12. 🤖 Smart Content Auto-Complete
**Impact**: Very High (AI-powered writing assistant)
In-line AI suggestions while typing experience bullets. User types partial text, presses Tab, and AI completes the sentence. Requires a streaming endpoint (`SSE` or `WebSocket`) that calls Gemini/OpenAI with the partial text + job title context. Most complex AI feature but highest engagement potential.

### 13. LinkedIn Import via Paste
**Impact**: High (saves 15-30 min of manual entry)
Add a "Paste LinkedIn Profile" modal on `/input`. User pastes their LinkedIn profile URL or raw text. Send it to AI with structured extraction prompt (similar to your existing PDF extraction). Maps LinkedIn sections → [CVData](file:///Users/sammidev/Developments/atstex-lab/internal/cvtemplate/cvtemplate.go#131-143) fields. Reuses `extractor.ExtractBiodata` with a modified prompt.

### 14. Email Notifications (Resend Integration)
**Impact**: Medium-High (re-engagement)
Integrate Resend (or SMTP) for:
- "Your application to X has been 'Applied' for 7+ days — time to follow up?"
- "Your subscription expires in 3 days"
- Weekly digest: "You have 3 pending applications"
Requires a background cron job (can use Go's `time.Ticker` or an external scheduler).

### 15. Multi-Language CV Content (i18n Templates)
**Impact**: Medium (international users)
Auto-translate static CV section headers ("Experience" → "Pengalaman", "Education" → "Pendidikan") based on a language selector. The LaTeX templates already use Go template functions — add a `t` function that looks up translations from a map.

### 16. Comprehensive Admin Analytics Dashboard
**Impact**: High (business intelligence)
Extend [AdminStats](file:///Users/sammidev/Developments/atstex-lab/internal/domain/user.go#30-37) with: daily active users, compilations per day, most popular templates, AI token burn rate vs. subscription revenue, error rates. Use Chart.js or similar on the admin page. Requires new repository queries with `DATE_TRUNC` aggregations.

---

## 📊 Priority Matrix

| Feature | Effort | Impact | ROI |
|---------|--------|--------|-----|
| Dark Mode | ⚡ Easy | ⭐⭐⭐ | 🔥🔥🔥 |
| CV Completeness Score | ⚡ Easy | ⭐⭐⭐ | 🔥🔥🔥 |
| AI Interview Prep | 🟡 Medium | ⭐⭐⭐⭐⭐ | 🔥🔥🔥🔥 |
| AI Bullet Enhancer | 🟡 Medium | ⭐⭐⭐⭐ | 🔥🔥🔥🔥 |
| JSON Export/Import | ⚡ Easy | ⭐⭐ | 🔥🔥🔥 |
| CV Version History | 🟡 Medium | ⭐⭐⭐ | 🔥🔥🔥 |
| Kanban Stats + Deadlines | 🟡 Medium | ⭐⭐⭐ | 🔥🔥🔥 |
| PDF Watermark Gate | 🟡 Medium | ⭐⭐⭐⭐ | 🔥🔥🔥🔥 |
| Template Color Picker | 🟡 Medium | ⭐⭐ | 🔥🔥 |
| LinkedIn Import | 🔴 Hard | ⭐⭐⭐⭐ | 🔥🔥🔥 |
| Smart Auto-Complete | 🔴 Hard | ⭐⭐⭐⭐⭐ | 🔥🔥🔥 |
| Admin Analytics | 🔴 Hard | ⭐⭐⭐ | 🔥🔥 |

---

> **My recommendation**: Start with **Dark Mode** + **CV Completeness Score** (both pure frontend, ship in a day), then tackle **AI Interview Prep** and **AI Bullet Enhancer** as your next high-impact features — they directly leverage your existing AI infrastructure and create real differentiation.
