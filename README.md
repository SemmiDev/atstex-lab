# ATSTEX-LAB

A modern, privacy-first LaTeX CV builder designed to help you generate beautiful, ATS-friendly resumes directly from your browser.

Write your biodata, choose a template, and instantly compile it into a professional PDF.

![Logo](/static/cv-favicon.svg)

---

## 🏗️ Architecture & Flow

Atstex-Lab acts as a bridge between a streamlined web frontend and the powerful LaTeX typesetting engine.

```mermaid
flowchart TB
    %% Users and external triggers
    User((User))

    %% Frontend Tier
    subgraph Client [Browser / Client Tier]
        direction TB
        UI[Web UI <br/>HTML/JS + Tailwind]
        Editor[CV Data Editor<br/>Local Storage]
        PreviewTool[PDF.js<br/>In-Browser Render]
        AuthUI[OAuth Login]

        UI <--> Editor
        PreviewTool -.-> UI
    end

    %% External Services
    subgraph External [External APIs & Services]
        direction LR
        GoogleAuth[Google OAuth Provider]
        OpenAI[AI Models<br/>OpenAI/Gemini/Anthropic]
    end

    %% Backend Server Tier
    subgraph Server [Go Backend API Tier]
        direction TB
        Router[HTTP Router &<br/>Middleware]

        subgraph Services [Business Logic]
            AuthSvc[Auth & Session<br/>Management]
            ProfileSvc[CV Profile<br/>Manager]
            AIEngine[AI Extraction Interface<br/>PDF -> Text -> JSON]
            TemplateSvc[LaTeX Template<br/>Injection & Routing]
        end

        subgraph Security [Security Layer]
            Sanitizer[Input Sanitizer<br/>Go String Escaping]
            Limits[OS Resource Limits<br/>Timeouts & Memory]
        end

        Compiler[Compilation Engine<br/>File I/O Coordinator]

        %% Internal routing
        Router --> AuthSvc
        Router --> ProfileSvc
        Router --> AIEngine
        Router --> TemplateSvc
        TemplateSvc --> Sanitizer
        Sanitizer --> Compiler
        Compiler --> Limits
    end

    %% Storage & Execution Infrastructure
    subgraph Infrastructure [Docker Infrastructure]
        direction TB
        DB[(PostgreSQL Database<br/>Sessions & CV Data)]

        subgraph Sandbox [Temporal Execution Sandbox]
            TempDir[/tmp/workspace<br/>Isolated Directory/]
            ImageDump[Temp Photo.jpg/png]
            TexSource[Generated document.tex]

            Tectonic([Tectonic TeX Engine<br/>Alpine/Cache Pre-loaded])

            ImageDump -.-> Tectonic
            TexSource -.-> Tectonic
        end
    end

    %% Data Flow Connections
    User == Fills Form / Uploads Photo ==> UI
    User == Signs In ==> AuthUI
    User == Uploads Resume PDF ==> UI

    AuthUI <-->|OAuth Tokens| GoogleAuth
    AuthUI <-->|Session Cookies| Router

    UI == Syncs Biodata JSON ==> Router
    UI == Manual Compile Request ==> Router
    UI == Extract PDF Request ==> Router

    AuthSvc <-->|Validates/Stores| DB
    ProfileSvc <-->|Saves/Loads JSON| DB

    AIEngine <-->|Sends Prompt & Text<br/>Receives JSON Resume| OpenAI

    Compiler == Writes Payload ==> TempDir
    Compiler == Writes Base64 Photo ==> ImageDump
    Compiler == Writes TeX Template ==> TexSource

    Limits == Executes Command<br/>with restricted privs ==> Tectonic
    Tectonic == Produces document.pdf ==> TempDir
    TempDir == Reads PDF Bytes ==> Compiler

    Compiler == Streams PDF Response ==> Router
    Router == Returns Application/PDF ==> PreviewTool

    %% Styling Elements
    classDef primary fill:#ff4794,stroke:#000,stroke-width:2px,color:#fff,font-weight:bold;
    classDef secondary fill:#475eff,stroke:#000,stroke-width:2px,color:#fff;
    classDef database fill:#5eeb8f,stroke:#000,stroke-width:2px,color:#000;
    classDef external fill:#f3f4f6,stroke:#000,stroke-width:2px,color:#000,stroke-dasharray: 5 5;
    classDef secure fill:#ffbd45,stroke:#000,stroke-width:2px,color:#000;
    classDef sandbox fill:#e2e8f0,stroke:#64748b,stroke-width:2px,stroke-dasharray: 5 5;

    class UI,Editor,PreviewTool,AuthUI primary;
    class Router,AuthSvc,ProfileSvc,AIEngine,TemplateSvc,Compiler secondary;
    class DB,Tectonic database;
    class GoogleAuth,OpenAI external;
    class Sanitizer,Limits secure;
    class TempDir,ImageDump,TexSource sandbox;
```

### How it Works

1. **Input**: You fill out your professional profile data using the Web UI, which saves securely to your persistent PostgreSQL database profile.
2. **AI Extraction**: Optionally, upload an existing PDF resume. The Go backend sends the text to an AI provider (like OpenAI or Gemini) to intelligently parse and auto-fill your profile.
3. **Compilation**: The backend engine marries your JSON payload with a LaTeX blueprint template, injects uploaded assets (like Profile Photos), and executes `tectonic` in an isolated temporal sandbox to compile the raw `document.tex` into a flawless PDF.

---

## 🚀 Quickstart using Docker

The absolute best way to run ATSTEX-LAB locally is via Docker. This bundles the Go API, PostgreSQL database, and the massive TeX Live compilation engine into a single containerized environment.

### 1. Setup Environment

Clone the repository, then copy the environment variables:

```bash
cp .env.example .env
```

*(Optional)* Add your `GOOGLE_CLIENT_ID` and `AI_API_KEY` to enable OAuth login and AI PDF parsing.

### 2. Run the Stack

Run the makefile command to build the image and start the database:

```bash
make docker-run
```

*Note: The first build will take a few minutes as it downloads and pre-caches the massive Tectonic package bundle.*

### 3. Open the App

Visit [http://localhost:8080](http://localhost:8080) in your browser!

### Common Commands

- `make docker-logs` - Tail the active output logs.
- `make docker-down` - Stop the containers safely.
- `make docker-remove-rebuild` - Nuke the database and container volumes to start fresh.

---

## 🛠️ Local Development (Without Docker)

If you'd like to develop natively without Docker containers, ensure you have **Go 1.22+**, **Node.js**, **PostgreSQL**, and a full LaTeX distribution installed on your machine (`mactex` on macOS or `texlive-full` on Linux).

```bash
# Compile CSS tailwind watchers and boot the Go development server
npm install
make run

# Build the standalone binary
make build
./atstex-lab
```
