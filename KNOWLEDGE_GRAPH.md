# ATSTEX-LAB — Complete AI Knowledge Graph

> **Purpose of this file:** A single-source-of-truth reference for any AI assistant (or new engineer) to instantly understand every layer of the ATSTEX-LAB codebase — architecture, data model, APIs, features, security, deployment, and inter-package relationships.

---

## 1. PROJECT IDENTITY

| Property       | Value                                                          |
|----------------|----------------------------------------------------------------|
| **Name**       | ATSTEX-LAB                                                     |
| **Purpose**    | LaTeX CV/resume builder + AI-powered career tools platform     |
| **Module**     | `github.com/semmidev/atstex-lab`                              |
| **Language**   | Go 1.24.4 (backend) + HTML/CSS/JS (frontend)                  |
| **Repository** | https://github.com/SemmiDev/atstex-lab                        |
| **License**    | ISC                                                            |
| **Entry Point**| `./cmd/server/main.go`                                         |
| **Binary**     | `./atstex-lab`                                                 |
| **Default Port**| `8080`                                                        |

### Core Value Proposition
Users fill out a structured biodata form → choose a LaTeX CV template → the backend sanitizes, renders, and compiles the `.tex` source via a LaTeX engine → returns a professional PDF. The same biodata also powers four AI tools: ATS scoring, CV critique, cover letter generation, and interview question generation.

---

## 2. HIGH-LEVEL ARCHITECTURE

```mermaid
flowchart TB
    User((Authenticated User))
    Viewer((Public Viewer))
    Admin((Admin User))

    subgraph Browser["Browser / Client Tier"]
        UI["Go HTML Templates + Tailwind CSS + Vanilla JS\n(editor.js, kanban.js, input.js, gallery.js, admin.js, i18n.js, dark-mode.js)"]
    end

    subgraph GoServer["Go Backend — internal/handler/Server"]
        Router["Chi Router\nMiddleware Stack:\n• RequestID • RealIP • Logger\n• Recoverer • Timeout(120s)\n• CleanPath • StripSlashes"]

        subgraph Features["Feature Handlers"]
            Auth["Auth Handler\nGoogle OAuth2 flow"]
            CVProfile["CV Profile Handler\nCRUD biodata"]
            Compiler["Compile Handler\nLaTeX → PDF"]
            AITools["AI Tool Handlers\n• CV Review\n• Cover Letter\n• ATS Simulator\n• Interview Prep"]
            Kanban["Job Application Handler\nKanban board CRUD"]
            PublicProfile["Public Profile Handler\npublic CV portfolios"]
            AdminH["Admin Handler\nuser mgmt + analytics"]
            SubMgmt["Subscription Handler\nplan + assignment mgmt"]
        end

        Router --> Features
    end

    subgraph Packages["Internal Go Packages"]
        AuthPkg["auth\nOAuth2 config + middleware"]
        ConfigPkg["config\nenv var loader"]
        DomainPkg["domain\nall entity structs"]
        RepoPkg["repository\nPostgres data layer"]
        ExtPkg["extractor\nLangChainGo AI wrapper"]
        CVTPkg["cvtemplate\ntemplate list/render/data"]
        CompPkg["compiler\nLaTeX engine runner"]
        MWPkg["middleware\nslog request logger"]
        WebPkg["web\nembed.FS for assets"]
    end

    subgraph Infra["External Infrastructure"]
        PG[("PostgreSQL 18\n11 tables")]
        AI["AI Providers\nOpenAI / Anthropic\nOllama / Gemini / Groq"]
        GoogleOAuth["Google OAuth2 API"]
    end

    User <-->|HTTP| Browser
    Viewer -.->|HTTP| Browser
    Admin <-->|HTTP| Browser
    Browser <-->|HTML + REST JSON| Router
    Features --> Packages
    RepoPkg --> PG
    ExtPkg --> AI
    Auth --> GoogleOAuth
```

---

## 3. DIRECTORY STRUCTURE MAP

```
atstex-lab/
├── cmd/
│   └── server/
│       └── main.go                  # Entry point: loads config, connects DB, starts server
│
├── internal/
│   ├── auth/
│   │   └── auth.go                  # Google OAuth2 config + 3 HTTP middleware funcs
│   ├── compiler/
│   │   ├── compiler.go              # LaTeX → PDF execution (pdflatex/xelatex/lualatex/tectonic)
│   │   ├── sanitize.go              # SanitizeLatexInput + SanitizeCVData (10 special chars)
│   │   ├── rlimit_linux.go          # OS resource limits (Linux only)
│   │   └── rlimit_other.go          # No-op stub for non-Linux
│   ├── config/
│   │   └── config.go                # AppConfig struct + Load() from .env
│   ├── cvtemplate/
│   │   └── cvtemplate.go            # TemplateInfo, CVData struct, PageSettings, Render()
│   ├── domain/
│   │   ├── user.go                  # User, AdminStats, AdminAnalytics, AdminUserRow
│   │   ├── session.go               # Session
│   │   ├── cv_profile.go            # CVProfile
│   │   ├── cv_review.go             # CVReview
│   │   ├── cover_letter.go          # CoverLetter
│   │   ├── ats_simulation.go        # AtsSimulation
│   │   ├── interview_prep.go        # InterviewPrep + InterviewPrepResult
│   │   ├── job_application.go       # JobApplication
│   │   ├── subscription.go          # SubscriptionPlan + UserSubscription
│   │   └── feedback.go              # Feedback
│   ├── extractor/
│   │   └── extractor.go             # AIConfig + LLM client factory + 4 AI functions
│   ├── handler/
│   │   ├── server.go                # Server struct + all page/API handlers (1273 lines)
│   │   ├── routes.go                # Chi route registration (public/auth/protected/admin)
│   │   ├── auth.go                  # Google login/callback/logout/delete-account
│   │   ├── cv_profile.go            # CV profile CRUD handlers
│   │   ├── ats_simulator.go         # ATS simulation page + API
│   │   ├── interview_prep.go        # Interview prep page + API
│   │   ├── job_application.go       # Kanban board CRUD handlers
│   │   ├── public_profile.go        # Public profile + PDF download + username mgmt
│   │   └── admin_subscription.go   # Admin subscription plan + assignment handlers
│   ├── middleware/
│   │   └── logger.go                # slog-based HTTP request logger middleware
│   └── repository/
│       ├── repository.go            # Repository interface (60 methods) + postgresRepo impl
│       ├── ats_simulation.go        # ATS simulation extra repo methods
│       └── interview_prep.go        # Interview prep extra repo methods
│
├── web/
│   ├── embed.go                     # embed.FS for templates + static assets
│   ├── templates/
│   │   ├── *.html                   # 17 HTML templates (Go template engine)
│   │   └── cv/
│   │       ├── bay.tex              # CV template: Bay
│   │       ├── delta.tex            # CV template: Delta
│   │       ├── reef.tex             # CV template: Reef
│   │       ├── sea.tex              # CV template: Sea
│   │       ├── tide.tex             # CV template: Tide
│   │       └── wave.tex             # CV template: Wave
│   └── static/
│       ├── css/                     # Tailwind + per-page CSS
│       └── js/                      # Per-page vanilla JS modules
│
├── deploy/                          # Ansible IaC for production Ubuntu server
│   ├── playbooks/                   # site.yml, setup.yml, deploy.yml
│   ├── roles/                       # common, docker, nginx, postgres, app
│   ├── group_vars/                  # all.yml, production.yml (Vault-encrypted)
│   └── scripts/                     # backup-db.sh, restore-db.sh, health-check.sh
│
├── init.sql                         # PostgreSQL schema + seed data
├── compose.yml                      # Docker Compose: postgres + atstex-lab services
├── Dockerfile                       # Multi-stage: Go builder → Alpine + Tectonic runtime
├── Makefile                         # run, build, css, docker-*, install-latex targets
├── go.mod                           # Go module dependencies
├── tailwind.config.js               # Tailwind CSS configuration
└── .env                             # Environment variables (not committed)
```

---

## 4. DOMAIN MODEL (Entity Relationship)

```mermaid
erDiagram
    users {
        uuid    id          PK
        string  google_id   UK
        string  email       UK
        string  name
        string  picture
        string  role        "user|admin"
        bigint  ai_tokens_used
        string  username    UK "nullable"
        bool    is_blocked
        time    created_at
        time    updated_at
    }

    sessions {
        uuid    id          PK
        uuid    user_id     FK
        string  token       UK
        string  ip_address
        string  user_agent
        time    expires_at
        time    created_at
    }

    cv_profiles {
        uuid    id          PK
        uuid    user_id     FK
        string  title
        jsonb   biodata     "CVData JSON blob"
        bool    is_public
        time    created_at
        time    updated_at
    }

    cv_reviews {
        uuid    id          PK
        uuid    user_id     FK
        uuid    profile_id  FK
        string  profile_title
        string  language    "en|id|..."
        int     score       "0-100"
        text    strengths
        text    improvements
        text    recommendations
        bigint  tokens_used
        time    created_at
    }

    cover_letters {
        uuid    id          PK
        uuid    user_id     FK
        uuid    profile_id  FK
        string  profile_title
        text    job_description
        text    cover_letter_text
        string  language
        bigint  tokens_used
        time    created_at
    }

    ats_simulations {
        uuid    id          PK
        uuid    user_id     FK
        uuid    profile_id  FK
        text    job_description
        int     score
        jsonb   missing_keywords "string[]"
        text    recommendations
        time    created_at
    }

    interview_preps {
        uuid    id          PK
        uuid    user_id     FK
        uuid    profile_id  FK
        text    job_description
        string  language
        jsonb   questions   "InterviewQuestionCategory[]"
        bigint  tokens_used
        time    created_at
    }

    job_applications {
        uuid    id            PK
        uuid    user_id       FK
        uuid    cv_profile_id FK "nullable"
        string  company
        string  job_title
        string  status        "Applied|Interviewing|Offered|Rejected|..."
        text    notes
        date    deadline      "nullable"
        time    created_at
        time    updated_at
    }

    feedbacks {
        uuid    id          PK
        uuid    user_id     FK
        string  subject
        text    message
        text    admin_reply "nullable"
        time    replied_at  "nullable"
        time    created_at
    }

    subscription_plans {
        uuid    id                  PK
        string  name                "Gratis|Basic|Pro"
        bigint  price_idr
        int     duration_months
        int     max_cv_profiles     "-1=unlimited"
        int     max_cv_reviews      "-1=unlimited"
        int     max_ats_simulations
        int     max_cover_letters
        bool    is_active
        time    created_at
        time    updated_at
    }

    user_subscriptions {
        uuid    id          PK
        uuid    user_id     FK
        uuid    plan_id     FK
        time    start_date
        time    end_date
        time    created_at
    }

    users ||--o{ sessions : "has"
    users ||--o{ cv_profiles : "owns"
    users ||--o{ cv_reviews : "generates"
    users ||--o{ cover_letters : "generates"
    users ||--o{ ats_simulations : "runs"
    users ||--o{ interview_preps : "generates"
    users ||--o{ job_applications : "tracks"
    users ||--o{ feedbacks : "sends"
    users ||--o{ user_subscriptions : "subscribes"
    cv_profiles ||--o{ cv_reviews : "reviewed-by"
    cv_profiles ||--o{ cover_letters : "source-for"
    cv_profiles ||--o{ ats_simulations : "scored-by"
    cv_profiles ||--o{ interview_preps : "basis-for"
    cv_profiles ||--o{ job_applications : "linked-to"
    subscription_plans ||--o{ user_subscriptions : "purchased-as"
```

---

## 5. GO PACKAGE DEPENDENCY GRAPH

```mermaid
graph TD
    main["cmd/server/main.go"]

    main --> config
    main --> repository
    main --> auth
    main --> handler
    main --> extractor

    handler --> auth
    handler --> domain
    handler --> repository
    handler --> compiler
    handler --> cvtemplate
    handler --> extractor
    handler --> middleware
    handler --> web

    auth --> domain
    auth --> repository

    compiler --> cvtemplate

    cvtemplate --> web

    extractor --> cvtemplate

    repository --> domain

    middleware["middleware\n(logger)"]

    style main fill:#ff4794,color:#fff
    style handler fill:#475eff,color:#fff
    style domain fill:#5eeb8f,color:#000
    style repository fill:#5eeb8f,color:#000
    style extractor fill:#f59e0b,color:#000
    style compiler fill:#ef4444,color:#fff
    style cvtemplate fill:#8b5cf6,color:#fff
```

---

## 6. SERVER STRUCT & DEPENDENCIES

```go
// internal/handler/server.go
type Server struct {
    router     *chi.Mux
    tmpl       *template.Template     // all 17 HTML templates, embedded at startup
    logger     *slog.Logger           // structured JSON logger
    repo       repository.Repository  // interface → postgresRepo implementation
    authConfig *auth.Config           // Google OAuth2 config
    aiConfig   extractor.AIConfig     // AI provider settings (provider/model/key/baseURL)
}
```

**Key Server methods:**
| Method | Purpose |
|--------|---------|
| `NewServer(cfg, repo, authCfg, aiCfg)` | Constructor, calls `routes()` and parses templates |
| `ServeHTTP(w, r)` | Delegates to `s.router` |
| `routes()` | Registers all Chi routes with middleware groups |
| `encode(w, r, status, v)` | JSON response helper |
| `respondError(w, r, err, status)` | Error response helper |
| `isFreeTier(ctx, userID)` | Checks if user has no active paid subscription |
| `checkSubscriptionLimits(ctx, userID, feature)` | Enforces plan quotas for AI features |

---

## 7. COMPLETE API ROUTE MAP

### 7.1 Public Routes (no auth required, optional user injection)
| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/` | `handleHome` | Landing page |
| GET | `/u/{username}` | `handlePublicProfile` | Public CV portfolio |
| GET | `/u/{username}/{profileID}/pdf` | `handlePublicProfileDownloadPDF` | Download public CV as PDF |

### 7.2 Authentication Routes
| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/auth/google/login` | `handleGoogleLogin` | Redirect to Google OAuth |
| GET | `/auth/google/callback` | `handleGoogleCallback` | OAuth callback, upsert user, create session |
| POST | `/auth/logout` | `handleLogout` | Clear session cookie + DB record |
| POST | `/auth/delete-account` | `handleDeleteAccount` | Soft-delete user + clear session |
| GET | `/forbidden` | `handleForbiddenPage` | Blocked user page |

### 7.3 Protected Routes (requires valid session cookie)
#### Pages
| Method | Path | Template | Description |
|--------|------|---------|-------------|
| GET | `/input` | `input` | Biodata input form |
| GET | `/input/embed` | `input_embed` | Embeddable input form |
| GET | `/editor` | `editor` | LaTeX preview editor |
| GET | `/gallery` | `gallery` | Multi-template gallery |
| GET | `/publish` | `publish` | Public profile publish settings |
| GET | `/profile` | `profile` | User profile + sessions |
| GET | `/subscription` | `subscription` | Subscription status page |
| GET | `/cv-review` | `cv-review` | AI CV critique page |
| GET | `/cover-letter` | `cover-letter` | AI cover letter page |
| GET | `/ats-simulator` | `ats-simulator` | ATS scoring page |
| GET | `/interview-prep` | `interview-prep` | AI interview prep page |
| GET | `/kanban` | `kanban` | Job application tracker |
| GET | `/feedback` | `feedback` | Feedback form page |

#### CV Profile API
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/cv-profiles` | List all user's CV profiles |
| POST | `/api/cv-profiles` | Create new CV profile |
| GET | `/api/cv-profiles/{id}` | Get single profile |
| PUT | `/api/cv-profiles/{id}` | Save biodata JSON |
| PUT | `/api/cv-profiles/{id}/title` | Rename profile |
| DELETE | `/api/cv-profiles/{id}` | Delete profile |
| PUT | `/api/cv-profiles/{id}/visibility` | Toggle public/private |

#### Template & Compilation API
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/templates` | List all .tex templates |
| GET | `/api/templates/{name}` | Get raw .tex source |
| POST | `/api/templates/{name}/render` | Render template with CVData → LaTeX string |
| POST | `/compile` | Compile LaTeX → PDF (base64) |
| POST | `/api/gallery/compile` | Compile all templates in parallel |
| POST | `/api/page-settings/apply` | Apply page settings to template |
| POST | `/api/extract-pdf` | Upload PDF → AI parse → biodata JSON |

#### AI Tools API
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/cv-review` | Generate AI CV critique → save |
| GET | `/api/cv-reviews` | List user's CV reviews |
| POST | `/api/cover-letter/generate` | Generate cover letter → save |
| GET | `/api/cover-letters` | List user's cover letters |
| POST | `/api/ats-simulator` | Score CV vs job description → save |
| GET | `/api/ats-simulations` | List user's ATS simulations |
| POST | `/api/interview-prep` | Generate interview questions → save |
| GET | `/api/interview-preps` | List user's interview preps |

#### Job Application / Kanban API
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/applications` | List all job applications |
| POST | `/api/applications` | Create job application |
| PUT | `/api/applications/{id}` | Update job application details |
| PUT | `/api/applications/{id}/status` | Update kanban status |
| DELETE | `/api/applications/{id}` | Delete job application |

#### Other Protected API
| Method | Path | Description |
|--------|------|-------------|
| PUT | `/api/username` | Set/update public username |
| GET | `/api/username/check` | Check username availability |
| GET | `/api/feedback` | List user's submitted feedbacks |
| POST | `/api/feedback` | Submit new feedback |
| POST | `/auth/sessions/{token}/delete` | Revoke a specific session |

### 7.4 Admin Routes (requires login + `role = "admin"`)
#### Admin Pages
| Method | Path | Description |
|--------|------|-------------|
| GET | `/admin` | Admin dashboard page |

#### Admin API
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/stats` | Aggregate platform statistics |
| GET | `/api/admin/analytics` | Time-series daily counts |
| GET | `/api/admin/users` | Paginated/searchable user list |
| GET | `/api/admin/feedbacks` | All user feedbacks |
| POST | `/api/admin/feedbacks/{id}/reply` | Reply to feedback |
| DELETE | `/api/admin/feedbacks/{id}` | Delete feedback |
| POST | `/api/admin/users/{id}/block` | Block user |
| POST | `/api/admin/users/{id}/unblock` | Unblock user |
| POST | `/api/admin/users/{id}/make-admin` | Grant admin role |
| POST | `/api/admin/users/{id}/revoke-admin` | Revoke admin role |
| DELETE | `/api/admin/users/{id}` | Delete user |
| GET | `/api/admin/subscription-plans` | List all subscription plans |
| POST | `/api/admin/subscription-plans` | Create new plan |
| PUT | `/api/admin/subscription-plans/{id}` | Update plan |
| DELETE | `/api/admin/subscription-plans/{id}` | Delete plan |
| POST | `/api/admin/subscription-plans/{id}/toggle` | Activate/deactivate plan |
| POST | `/api/admin/users/{id}/subscribe` | Assign subscription to user |
| GET | `/api/admin/users/{id}/subscription` | Get user's active subscription |

---

## 8. AUTHENTICATION & SESSION FLOW

```mermaid
sequenceDiagram
    participant Browser
    participant Server
    participant GoogleAPI
    participant PostgreSQL

    Browser->>Server: GET /auth/google/login
    Server->>Server: GenerateStateOauthCookie() → set cookie "oauthstate"
    Server->>Browser: 307 Redirect → Google OAuth URL

    Browser->>GoogleAPI: User grants permission
    GoogleAPI->>Browser: 302 Redirect to /auth/google/callback?code=...&state=...

    Browser->>Server: GET /auth/google/callback
    Server->>Server: Validate state cookie
    Server->>GoogleAPI: Exchange code → access_token
    Server->>GoogleAPI: GET /oauth2/v2/userinfo
    GoogleAPI->>Server: {id, email, name, picture}

    Server->>PostgreSQL: UpsertUser(googleID, email, name, picture)
    PostgreSQL->>Server: User record

    Server->>PostgreSQL: GetUserActiveSubscription(userID)
    alt No subscription found
        Server->>PostgreSQL: GetFreeSubscriptionPlan()
        Server->>PostgreSQL: AdminAssignSubscription(userID, freePlanID, months)
    end

    Server->>Server: GenerateSessionToken() → 32-byte random base64
    Server->>PostgreSQL: CreateSession(userID, token, ip, ua, expires=30days)
    Server->>Browser: Set-Cookie: session_token=... (HttpOnly, SameSite=Lax, 30d)
    Browser->>Server: 307 Redirect → /admin (if admin) or /input (if user)
```

### Middleware Chain (per protected request)
```
Request
  → chimw.RequestID        (inject X-Request-ID)
  → chimw.RealIP           (extract real IP)
  → mw.RequestLogger       (slog structured log + inject logger into ctx)
  → chimw.Recoverer        (panic recovery)
  → chimw.Timeout(120s)    (request deadline)
  → chimw.CleanPath        (normalize URL paths)
  → chimw.StripSlashes     (remove trailing slashes)
  → auth.Middleware        (validate session cookie → load User into ctx)
    [→ auth.AdminMiddleware (check role == "admin", protected admin group only)]
  → Handler
```

### Context Keys
| Key | Type | Set By | Used For |
|-----|------|--------|----------|
| `auth.UserContextKey ("user")` | `*domain.User` | `auth.Middleware` | Access current user in handlers |
| `middleware.RequestIDKey` | `string` | `RequestLogger` | Request correlation |
| `middleware.TraceIDKey` | `string` | `RequestLogger` | Distributed trace ID |
| `middleware.LoggerKey` | `*slog.Logger` | `RequestLogger` | Per-request structured logger |

---

## 9. LATEX COMPILATION PIPELINE

```mermaid
flowchart LR
    A["User submits\nbiodata JSON\n+ engine choice\n+ optional photo"] --> B

    B["handler.handleCompile()\n1. Decode JSON body\n2. Auth check\n3. Fetch cv_profiles"]

    B --> C["compiler.SanitizeCVData()\nEscape 10 LaTeX special chars:\n\\ { } $ & # ^ _ ~ %\nin ALL string fields of CVData"]

    C --> D["cvtemplate.Render(name, cvData, pageSettings)\n1. Get .tex template from embed.FS\n2. Go template engine with [[ ]] delims\n3. Inject CVData + PageSettings\n4. Optional watermark injection"]

    D --> E["compiler.Compile(ctx, source, opts)\n1. Validate source size ≤ 512KB\n2. exec.LookPath(engine) check\n3. os.MkdirTemp() → workDir\n4. Write document.tex\n5. Decode + write photo.jpg/png"]

    E --> F{"Engine?"}

    F -->|tectonic| G["tectonic --outdir workDir\n--untrusted document.tex\n(auto multi-pass)"]
    F -->|pdflatex\nxelatex\nlualatex| H["engine -no-shell-escape\n-interaction=nonstopmode\n-halt-on-error\n-file-line-error\n(2 passes unless SinglePass)"]

    G --> I["Read document.pdf\ncleanLog()"]
    H --> I

    I --> J["Return Result{PDF, Log, Elapsed, Engine}"]
    J --> K["handler: base64 encode PDF\nReturn JSON {ok, pdfBase64, elapsed, engine}"]

    style C fill:#ef4444,color:#fff
    style E fill:#f59e0b,color:#000
    style G fill:#8b5cf6,color:#fff
    style H fill:#8b5cf6,color:#fff
```

### Supported LaTeX Engines
| Engine | Notes |
|--------|-------|
| `pdflatex` | Default. Classic TeX engine, fast, broad compatibility |
| `xelatex` | Unicode + OpenType font support |
| `lualatex` | Lua scripting + Unicode fonts |
| `tectonic` | **Primary Docker engine.** Auto-downloads packages, smaller image, single-pass capable |

### Security Measures in Compilation
- `-no-shell-escape` — prevents `\write18` shell injection
- `-interaction=nonstopmode` — no interactive prompts
- `maxSourceBytes = 512KB` — source size limit
- `defaultTimeout = 30s` — per-compilation deadline
- `setProcessLimits(cmd)` — OS-level rlimit (Linux: CPU time, memory, file size)
- `os.MkdirTemp()` + `defer os.RemoveAll(workDir)` — isolated temp sandbox, auto-cleaned
- `SanitizeCVData()` — escapes all 10 LaTeX special characters in user input

---

## 10. AI INTEGRATION LAYER

### extractor.AIConfig
```go
type AIConfig struct {
    Provider string // "openai" (default) | "anthropic" | "ollama" | "gemini"
    Model    string // "gpt-4o-mini" | "claude-3-haiku-20240307" | "llama3" | ...
    APIKey   string
    BaseURL  string // optional: Groq, Together.ai, or any OpenAI-compatible endpoint
}
```

### Four AI Functions
| Function | Input | Output | Quota Feature Key |
|----------|-------|--------|-------------------|
| `ExtractBiodata(ctx, pdfText, cfg)` | Raw PDF text | `CVData` JSON | _(PDF upload, no quota)_ |
| `CritiqueCVProfile(ctx, biodataJSON, lang, cfg)` | CV biodata + language | `CVCritiqueResult{Score, Strengths, Improvements, Recommendations}` | `"cv_review"` |
| `ScoreATS(ctx, biodataJSON, jobDesc, lang, cfg)` | CV biodata + job description + language | `ATSCritiqueResult{Score, MissingKeywords[], Recommendations}` | `"ats_simulation"` |
| `GenerateCoverLetter(ctx, biodataJSON, jobDesc, tone, maxParagraphs, lang, cfg)` | CV + job + tone + paragraphs + language | `string` (cover letter text) | `"cover_letter"` |
| `GenerateInterviewQuestions(ctx, biodataJSON, jobDesc, lang, count, cfg)` | CV + job + language + count | `InterviewPrepResult{Categories[]}` | _(no quota currently)_ |

### AI Provider Selection (`extractor.newLLM`)
```
Provider = "openai" (default) → openai.New(token, model, [baseURL])
Provider = "anthropic"        → anthropic.New(token, model)
Provider = "ollama"           → ollama.New(serverURL, model)
Provider = "gemini"           → googleai.New(ctx, apiKey, model)
Provider = "vertexai"         → vertexai.New(ctx, project, location, model)
Provider = "mistral"          → mistral.New(apiKey, model)
```
LangChainGo (`github.com/tmc/langchaingo`) is used as the unified LLM abstraction layer. Token usage is extracted from LLM response metadata and persisted to `users.ai_tokens_used`.

### AI Flow per Feature
```mermaid
sequenceDiagram
    participant H as Handler
    participant E as extractor pkg
    participant LLM as AI Provider API
    participant DB as PostgreSQL

    H->>H: checkSubscriptionLimits(ctx, userID, feature)
    H->>DB: GetCVProfile(profileID) → biodata JSON
    H->>E: ScoreATS / CritiqueCVProfile / GenerateCoverLetter / GenerateInterviewQuestions
    E->>LLM: Chat completion with structured system prompt
    LLM->>E: JSON response string
    E->>E: Parse JSON → typed result struct
    E->>H: (result, tokensUsed, error)
    H->>DB: IncrementAITokensUsed(userID, tokensUsed)
    H->>DB: Create[Feature]Record(...)
    H->>H: JSON encode response → 201 Created
```

---

## 11. SUBSCRIPTION & QUOTA SYSTEM

### Default Plans (seeded in `init.sql`)
| Plan | Price (IDR) | Max CV Profiles | Max CV Reviews | Max ATS Simulations | Max Cover Letters |
|------|-------------|-----------------|----------------|---------------------|-------------------|
| **Gratis** | 0 | 1 | 5 | 5 | 10 |
| **Basic** | 20,000 | 3 | 10 | 10 | 10 |
| **Pro** | 30,000 | -1 (unlimited) | 50 | 50 | 50 |

### Quota Check Logic (`Server.checkSubscriptionLimits`)
```
feature = "cv_profile" | "cv_review" | "ats_simulation" | "cover_letter"

1. GetUserActiveSubscription(userID)
   → if ErrNotFound → treat as free tier (isFreeTier=true)

2. For each feature, COUNT existing records for current month/period
3. If count >= plan.MaxFeature (and max != -1) → return quota error

Feature-specific counting:
  "cv_profile"     → CountCVProfilesByUserID (total, not monthly)
  "cv_review"      → CountCVReviewsByDate (current month)
  "ats_simulation" → CountAtsSimulationsByDate (current month)
  "cover_letter"   → CountCoverLettersByDate (current month)
```

### Subscription Assignment Flow
- On first login: `GetFreeSubscriptionPlan()` → auto-assign Gratis plan for 1 month
- Admin can override: `POST /api/admin/users/{id}/subscribe` with `{planId, months}`
- `user_subscriptions` stores start/end dates; active = `end_date > NOW()`

---

## 12. CVDATA STRUCT (The Core Data Model)

```go
// internal/cvtemplate/cvtemplate.go
type CVData struct {
    Personal       Personal        // name, title, email, phone, location, linkedin, github, website
    Summary        string          // professional summary paragraph
    Experience     []Experience    // company, title, location, dates, bullets (newline-separated)
    Education      []Education     // institution, degree, location, dates, gpa, activities
    Projects       []Project       // name, role, link, bullets
    Skills         Skills          // languages, frameworks, tools, other (comma-separated strings)
    Certifications []Certification // name, issuer
    Volunteer      []Volunteer     // organization, role, location, dates, bullets
    Awards         []Award         // title, issuer, date, description
    Talks          []Talk          // title, event, location, date, description
}

type PageSettings struct {
    DocumentClass string   // "article" (default), "report", "letter", "book"
    PaperSize     string   // "a4paper" (default), "letterpaper", "legalpaper"
    FontSize      string   // "10pt" (default), "11pt", "12pt"
    FontFamily    string   // "default" (Computer Modern), "helvet", "times", "palatino", "courier"
    MarginTop     string   // "0.60in" default
    MarginBottom  string   // "0.55in" default
    MarginLeft    string   // "0.70in" default
    MarginRight   string   // "0.70in" default
    LineSpacing   float64  // 1.0 (single), 1.5 (one-half), 2.0 (double)
    Alignment     string   // "justify" (default), "left", "center"
    HeaderText    string   // optional custom header
    FooterText    string   // optional custom footer
}
```

**Template rendering uses Go's `text/template` with `[[ ]]` delimiters** (instead of default `{{ }}`), to avoid conflicts with LaTeX braces. Custom template functions:

| Function | Purpose |
|----------|---------|
| `texEscape` | Escape LaTeX special chars in output |
| `texLines` | Split bullet strings by `\n` → `[]string` |
| `fmtSpacing` | Float → `\singlespacing` / `\onehalfspacing` / `\doublespacing` |
| `fontPkg` | Font family name → `\usepackage{...}` LaTeX command |
| `sprintf` | `fmt.Sprintf` wrapper |
| `ne` | Not-equal comparison |

---

## 13. CV TEMPLATES

All templates are LaTeX `.tex` files embedded in the binary via `web/embed.go`. They live in `web/templates/cv/`.

| Template | Description |
|----------|-------------|
| `bay.tex` | Template "Bay" style |
| `delta.tex` | Template "Delta" style |
| `reef.tex` | Template "Reef" style |
| `sea.tex` | Template "Sea" style |
| `tide.tex` | Template "Tide" style |
| `wave.tex` | Template "Wave" style |

Each template uses `[[ .Field ]]` syntax to inject `RenderData` (which embeds `CVData` + `PageSettings`). A watermark footer (`Built with ATSTEXLAB — atstexlab.com`) is injected before `\end{document}` when `showWatermark=true` (used for public PDF downloads).

---

## 14. FRONTEND TEMPLATES & JAVASCRIPT

### HTML Templates (Go `text/template`, embedded via `web/embed.go`)
| Template Name | File | Description |
|---------------|------|-------------|
| `home` | `home.html` | Landing page |
| `input` | `input.html` | Biodata form (main data entry) |
| `input_embed` | `input_embed.html` | Embeddable biodata form |
| `editor` | `editor.html` | Split-pane LaTeX editor + PDF preview |
| `gallery` | `gallery.html` | Multi-template side-by-side gallery |
| `cv-review` | `cv-review.html` | AI CV critique UI |
| `cover-letter` | `cover-letter.html` | AI cover letter generator UI |
| `ats-simulator` | `ats-simulator.html` | ATS score UI |
| `interview-prep` | `interview-prep.html` | AI interview question generator |
| `kanban` | `kanban.html` | Job application Kanban board |
| `profile` | `profile.html` | User profile + active sessions |
| `publish` | `publish.html` | Public profile publish settings |
| `public_profile` | `public_profile.html` | Public-facing CV portfolio |
| `subscription` | `subscription.html` | Subscription plan status |
| `feedback` | `feedback.html` | User feedback form |
| `admin` | `admin.html` | Admin dashboard |
| `forbidden` | `forbidden.html` | Blocked user page |

### JavaScript Modules (`web/static/js/`)
| File | Responsibilities |
|------|-----------------|
| `editor.js` | Template selector, CVData JSON ↔ PDF preview, compile call, photo upload, page settings panel |
| `input.js` | Dynamic biodata form: add/remove experience, education, projects, etc.; save to API |
| `kanban.js` | Drag-and-drop Kanban board, status column updates, CRUD modals |
| `gallery.js` | Fetch + display all templates compiled in parallel via `/api/gallery/compile` |
| `admin.js` | User management table, feedback replies, subscription assignment, analytics charts |
| `i18n.js` | Internationalization helper for UI string translations |
| `dark-mode.js` | Dark mode toggle with localStorage persistence |

### CSS Files (`web/static/css/`)
| File | Scope |
|------|-------|
| `tailwind.css` | Tailwind CSS source (compiled → `style.css`) |
| `style.css` | Compiled/minified Tailwind output |
| `editor.css` | Split-pane editor layout styles |
| `input.css` | Biodata form styles |
| `kanban.css` | Kanban board column + card styles |
| `gallery.css` | Gallery grid layout styles |
| `admin.css` | Admin dashboard table + chart styles |
| `home.css` | Landing page hero styles |

---

## 15. CONFIGURATION

### Environment Variables (`.env` / system env)
| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server listen port |
| `DATABASE_URL` | `postgres://atstex:password@localhost:5432/atstex?sslmode=disable` | PostgreSQL DSN |
| `GOOGLE_CLIENT_ID` | _(empty)_ | Google OAuth2 App client ID |
| `GOOGLE_CLIENT_SECRET` | _(empty)_ | Google OAuth2 App client secret |
| `GOOGLE_CALLBACK_URL` | `http://localhost:8080/auth/google/callback` | OAuth2 redirect URI |
| `AI_PROVIDER` | `openai` | AI provider: `openai`, `anthropic`, `ollama`, `gemini`, `vertexai`, `mistral` |
| `AI_MODEL` | `gpt-4o-mini` | Model identifier for the chosen provider |
| `AI_API_KEY` | _(empty)_ | API key (falls back to `OPENAI_API_KEY`) |
| `AI_BASE_URL` | _(empty)_ | Optional custom base URL for OpenAI-compatible endpoints |

---

## 16. DEPLOYMENT & INFRASTRUCTURE

### Docker Compose (`compose.yml`)
```
Services:
  postgres    → postgres:18-alpine
                volumes: pgdata, ./init.sql (init script)
                port: 5432

  atstex-lab  → built from Dockerfile
                port: 8080
                env_file: .env
                depends_on: postgres
                tmpfs: /tmp (512MB) — LaTeX temp compilation sandbox
                volumes: tectonic-cache (/var/cache/tectonic)
                resource limits: 2 CPUs, 1GB RAM
```

### Dockerfile (Multi-Stage)
```
Stage 1 (builder): golang:1.26-alpine
  → go mod download
  → CGO_ENABLED=0 go build -ldflags="-s -w" → /build/atstex-lab

Stage 2 (runtime): alpine:3.21
  → apk: tectonic, libstdc++, fontconfig, freetype
  → Pre-warm Tectonic package cache (dummy.tex compilation)
  → Non-root user: atstex-lab (uid 1001)
  → COPY binary from builder
  → EXPOSE 8080
  → ENTRYPOINT ["/usr/local/bin/atstex-lab"]
```

### Ansible IaC (`deploy/`)
```
Roles:
  common   → OS hardening: UFW firewall (ports 22/80/443), Fail2Ban, SSH key-only auth,
             swap file, unattended-upgrades, deploy user
  docker   → Install Docker Engine + Compose plugin
  nginx    → Nginx reverse proxy config + Let's Encrypt SSL (Certbot auto-renewal)
  postgres → PostgreSQL container setup, DB/user creation
  app      → Build Docker image, zero-downtime container swap, rollback tagging

Playbooks:
  site.yml   → Full server provision + deploy (first-time setup)
  setup.yml  → Infrastructure only (no app deploy)
  deploy.yml → App-only update (zero-downtime)

Scripts (installed to /opt/atstex-lab/scripts/ on server):
  backup-db.sh    → pg_dump → .sql.gz, Telegram notification, daily cron
  restore-db.sh   → Restore from .sql.gz backup file
  health-check.sh → Docker/Nginx/PostgreSQL status + disk/memory check
```

---

## 17. SECURITY MODEL

```mermaid
graph TD
    A["Incoming Request"] --> B["Chi Middleware Stack"]
    B --> C{Route Group?}

    C -->|Public| D["OptionalUserMiddleware\nTry to load user from cookie\nnon-blocking — proceeds without user"]
    C -->|Protected| E["auth.Middleware\nMUST have valid session_token cookie\nCheck session not expired\nLoad user from DB\nBlock check → redirect /forbidden"]
    C -->|Admin| F["auth.Middleware\n+\nauth.AdminMiddleware\nuser.Role must == 'admin'"]
    C -->|Auth endpoints| G["No auth middleware\n(public OAuth flow)"]

    D --> H["Handler (user may be nil)"]
    E --> I["Handler (user guaranteed non-nil)"]
    F --> J["Admin Handler (admin user guaranteed)"]

    style E fill:#ef4444,color:#fff
    style F fill:#7c3aed,color:#fff
    style D fill:#059669,color:#fff
```

### Security Checklist
| Concern | Mitigation |
|---------|-----------|
| **LaTeX Injection** | `SanitizeCVData()` escapes all 10 special chars before template render |
| **Shell Injection** | `-no-shell-escape` flag on all LaTeX engine invocations |
| **Resource Abuse** | 512KB source limit, 30s timeout, OS rlimits (Linux), 1GB RAM / 2 CPU Docker limits |
| **Session Security** | HttpOnly cookie, SameSite=Lax, 30-day expiry, server-side invalidation |
| **OAuth State** | CSRF state cookie validated before code exchange |
| **Blocked Users** | `is_blocked` check on every authenticated request → redirect `/forbidden` |
| **Profile Ownership** | Every profile handler verifies `profile.UserID == user.ID` |
| **Admin Escalation** | `AdminMiddleware` checks `role == "admin"` in context after `Middleware` |
| **SQL Injection** | All queries use parameterized statements via `sqlx` |
| **HTTPS** | Nginx + Let's Encrypt (production), UFW blocks all but 22/80/443 |
| **SSH Hardening** | Password auth disabled, root login disabled, Fail2Ban on port 22 |
| **DB Isolation** | PostgreSQL port 5432 bound to internal Docker network only |

---

## 18. REPOSITORY INTERFACE SUMMARY

The `repository.Repository` interface (60 methods) is the single data-access contract. The only implementation is `postgresRepo` using `sqlx` + `pgx/v5`.

```
User Management:    UpsertUser, GetUser, DeleteUser, SoftDeleteUser
Session:            CreateSession, GetSession, GetSessionsByUserID, DeleteSession
CV Profile:         CreateCVProfile, GetCVProfile, GetCVProfilesByUserID,
                    UpdateCVProfileBiodata, UpdateCVProfileTitle, DeleteCVProfile,
                    UpdateCVProfileVisibility, GetPublicCVProfilesByUserID
Username:           SetUsername, GetUserByUsername, CheckUsernameAvailable
AI Token Tracking:  IncrementAITokensUsed
Feedback:           CreateFeedback, GetFeedbacksByUserID,
                    AdminListFeedbacks, AdminReplyFeedback, AdminDeleteFeedback
CV Review:          CreateCVReview, GetCVReviewsByUserID, CountCVReviewsByDate
Cover Letter:       CreateCoverLetter, GetCoverLettersByUserID, CountCoverLettersByDate
Job Application:    CreateJobApplication, UpdateJobApplication, GetJobApplicationsByUserID,
                    UpdateJobApplicationStatus, DeleteJobApplication
ATS Simulation:     CreateAtsSimulation, GetAtsSimulationsByUserID, CountAtsSimulationsByDate
Interview Prep:     CreateInterviewPrep, GetInterviewPrepsByUserID, CountInterviewPrepsByDate
Admin User Mgmt:    AdminBlockUser, AdminUnblockUser, AdminDeleteUser,
                    AdminMakeUserAdmin, AdminRevokeUserAdmin,
                    AdminGetStats, AdminGetAnalytics, AdminListUsers
Subscription:       AdminListSubscriptionPlans, AdminCreateSubscriptionPlan,
                    AdminUpdateSubscriptionPlan, AdminToggleSubscriptionPlan,
                    AdminDeleteSubscriptionPlan, AdminAssignSubscription,
                    GetUserActiveSubscription, GetUserSubscriptions, GetFreeSubscriptionPlan
```

All PKs use UUIDv7 (`uuidv7()` PostgreSQL function) for time-sortable unique IDs.

---

## 19. KEY DATA FLOWS (End-to-End)

### Flow A: New User First Login
```
Browser → /auth/google/login
  → Google OAuth consent screen
  → /auth/google/callback
  → UpsertUser (INSERT ON CONFLICT UPDATE)
  → GetUserActiveSubscription → ErrNotFound
  → GetFreeSubscriptionPlan → AdminAssignSubscription (Gratis, 1 month)
  → CreateSession (30-day token)
  → Set-Cookie: session_token
  → Redirect /input
```

### Flow B: CV Compilation
```
Browser (editor.js) → POST /compile {source, engine, photoBase64}
  → auth.Middleware (validate session)
  → handler.handleCompile
  → SanitizeCVData(cvData)
  → cvtemplate.Render(templateName, cvData, pageSettings)
  → compiler.Compile(ctx, texSource, {engine, timeout, photoBase64})
    → os.MkdirTemp → write document.tex
    → [optional] decode + write photo.jpg
    → exec engine (2 passes or tectonic single)
    → read document.pdf
    → os.RemoveAll(workDir)
  → base64(pdf) → JSON response
  → editor.js renders PDF in <iframe>
```

### Flow C: AI ATS Scoring
```
Browser (ats-simulator page) → POST /api/ats-simulator {profileId, jobDescription, language}
  → auth.Middleware
  → handler.handleCreateAtsSimulation
  → GetCVProfile(profileID) → verify UserID ownership
  → checkSubscriptionLimits(userID, "ats_simulation")
  → extractor.ScoreATS(ctx, biodata, jobDesc, lang, aiConfig)
    → newLLM(ctx, cfg) → LangChainGo client
    → llms.GenerateFromSinglePrompt(systemPrompt + biodataJSON + jobDesc)
    → Parse JSON → ATSCritiqueResult{Score, MissingKeywords, Recommendations}
  → IncrementAITokensUsed(userID, tokensUsed)
  → CreateAtsSimulation(sim record)
  → JSON 201 Created
```

### Flow D: Public CV Portfolio
```
Browser → GET /u/{username}
  → OptionalUserMiddleware (try load viewer from cookie)
  → GetUserByUsername(username) → profileUser
  → GetPublicCVProfilesByUserID(profileUser.ID) → public profiles only
  → Render public_profile template with enriched profile data
  → (optional) GET /u/{username}/{profileID}/pdf
    → Verify profile.IsPublic == true
    → Parse biodata → CVData
    → cvtemplate.Render (with watermark=true)
    → compiler.Compile → PDF bytes
    → Content-Type: application/pdf + Content-Disposition: attachment
```

---

## 20. TECHNOLOGY STACK SUMMARY

```mermaid
mindmap
  root((ATSTEX-LAB))
    Backend
      Go 1.24.4
      Chi v5 Router
      sqlx + pgx/v5
      LangChainGo
      golang.org/x/oauth2
      slog structured logging
      embed.FS binary assets
    Frontend
      Go html/template
      Tailwind CSS v3
      Vanilla JavaScript ES6
      No frontend framework
    Database
      PostgreSQL 18
      UUIDv7 primary keys
      JSONB for biodata
      11 tables
    AI Providers
      OpenAI (default)
      Anthropic Claude
      Google Gemini / VertexAI
      Ollama (local)
      Mistral
      Any OpenAI-compatible API
    LaTeX Engines
      Tectonic (Docker default)
      pdflatex
      xelatex
      lualatex
    Infrastructure
      Docker + Compose
      Alpine Linux (runtime)
      Ansible IaC
      Nginx reverse proxy
      Let's Encrypt SSL
      UFW + Fail2Ban
    CV Templates
      bay.tex
      delta.tex
      reef.tex
      sea.tex
      tide.tex
      wave.tex
```

---

*This knowledge graph was auto-generated from full static analysis of the ATSTEX-LAB codebase. Update this file whenever significant architectural changes are made.*