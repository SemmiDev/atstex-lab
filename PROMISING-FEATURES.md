# Promising Features for the Next Sprint

Based on the current architecture of **atstex-lab** (ATS-friendly LaTeX CV generator with AI integration, Neo-Brutalist design, and multi-profile support), here is a curated list of high-impact features to consider for the next sprint. These features are categorized by their primary value proposition.

### B. ATS Simulator & Job Match Scoring
* **What it is:** Extending the recent "CV Critique" feature. Users paste a Job Description, and the AI acts as an ATS (Applicant Tracking System), scoring their CV match percentage and highlighting missing keywords they should add.
* **Why it's good:** Directly aligns with the core value proposition of being an "ATS-friendly" platform. It gives users actionable insights.
* **Effort:** Medium (Prompt engineering, new UI tab for comparison).

### C. Auto-Escaping Special Characters for LaTeX
* **What it is:** A background pre-processor that automatically escapes LaTeX special characters (like `&`, `%`, `$`, `#`, `_`) in user input to prevent compilation errors.
* **Why it's good:** Drastically improves user experience and compilation success rates. Many non-technical users get confused when their CV fails to compile due to a rogue `&` symbol.
* **Effort:** Low (Regex/string replacement utility before rendering).

---

### B. Rich Text / WYSIWYG Editor for Biodata
* **What it is:** Upgrade the textarea inputs for "Experience" and "Summary" to support a simple rich-text editor (bold, italic, bullet lists) that maps automatically to LaTeX commands (`\textbf{}`, `\textit{}`, `\begin{itemize}`).
* **Why it's good:** Makes data entry much more intuitive. Users won't need to know LaTeX to format their bullet points nicely.
* **Effort:** Medium (Integrating a lightweight JS editor like Quill/TipTap and writing a parser to LaTeX format).

### C. Import from LinkedIn
---

## 🛠 3. Professional & Power-User Tools

### A. Application Tracking Kanban Board
* **What it is:** A simple Kanban board (e.g., "Applied", "Interview", "Offer", "Rejected") where users can track the jobs they've applied for and link which CV profile they used for each.
* **Why it's good:** Turns the application from a "use once and leave" tool into a daily-use hub for job seekers.
* **Effort:** High (New DB tables, drag-and-drop UI component).

### B. Multi-Language CV Support (i18n)
* **What it is:** A feature to instantly auto-translate a CV profile into another language using AI, and generate the PDF in that language.
* **Why it's good:** Great for users applying to international jobs.
* **Effort:** Medium (AI translation workflow, ensuring LaTeX templates support unicode/specific language packages).

---

> [!TIP]
> **Recommendation:** I highly suggest starting with **Tailored Cover Letter Generator** or **ATS Simulator & Job Match Scoring**. Since you have already laid the groundwork with the AI integration and CV Critique features, these will offer the highest ROI with relatively low implementation effort.
