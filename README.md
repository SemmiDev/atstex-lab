# ATSTEX-LAB

Tulis biodata Anda, pilih *template* LaTeX, lalu *compile* menjadi dokumen PDF.

---

## Apa yang Dilakukannya?

- Isi data *resume* Anda melalui *form* penduan (informasi pribadi, pengalaman, pendidikan, keahlian, dll).
- Simpan beberapa profil CV sekaligus — satu untuk "Back End Developer", profil lain untuk "Data Science", dan seterusnya.
- Pilih dari pustaka *template* LaTeX yang ramah sistem ATS (*ATS-friendly*).
- *Compile* ke format PDF secara instan (mendukung *pdflatex*, *xelatex*, atau *lualatex*).
- Pratinjau hasil PDF Anda berdampingan dengan *form* pengisian biodata.
- Masuk / *Sign in* menggunakan Google untuk menyimpan data Anda tersinkronisasi di berbagai perangkat.

---

## Prasyarat

| Perangkat (Tool) | Fungsi (Why) | Instalasi |
|------------------|--------------|-----------|
| **Docker + Docker Compose** | Menjalankan aplikasi, PostgreSQL, dan TeX Live di dalam *container* | [docker.com](https://docs.docker.com/get-docker/) |
| **Go 1.22+** | Hanya dibutuhkan untuk *development* lokal (tidak diperlukan jika memakai Docker) | [go.dev](https://go.dev/dl/) |
| **Node.js + npm** | Melakukan kompilasi untuk Tailwind CSS | [nodejs.org](https://nodejs.org/) |

---

## Menjalankan dengan Docker (Direkomendasikan)

Ini adalah metode yang paling direkomendasikan dan paling mudah. *Container* aplikasi telah dilengkapi dengan instalasi TeX Live secara penuh, sehingga Anda tidak perlu repot menginstal komponen LaTeX tambahan di laptop Anda.

### 1. Mengatur Variabel Lingkungan (*Environment Variables*)

Salin dan ubah konfigurasi pada file `.env`:

```bash
cp .env.example .env
```

Lengkapi pengaturan koneksinya:

```env
DATABASE_URL=postgres://postgres:postgres@postgres:5432/atstex_lab?sslmode=disable
GOOGLE_CLIENT_ID=client-id-google-anda
GOOGLE_CLIENT_SECRET=rahasia-client-google-anda
GOOGLE_CALLBACK_URL=http://localhost:8080/auth/google/callback

# Konfigurasi Ekstraksi AI
# Provider yang didukung: openai (default), anthropic, googleai (gemini), mistral, ollama
AI_PROVIDER=openai
AI_MODEL=gpt-4o-mini
AI_API_KEY=api-key-anda
# AI_BASE_URL=        # Opsional: Untuk API yang kompatibel dengan format OpenAI (contoh: Groq, Together)
```

> Untuk mendapatkan kredensial Google OAuth, masuk ke [Google Cloud Console](https://console.cloud.google.com/apis/credentials) lalu buat sebuah *OAuth 2.0 Client ID*.

### 2. Memulai Aplikasi

```bash
make docker-run
```

Perintah ini akan menyusun (*build*) *image* Docker aplikasi lalu langsung menjalankan aplikasi beserta *database* PostgreSQL. *Catatan: Proses pertama kali akan memakan unduhan sekitar ~4 GB untuk dependensi basis gambar TeX Live.*

Buka browser dan arahkan ke **http://localhost:8080**.

### Perintah Docker Lainnya

```bash
make docker-logs            # Memantau log aplikasi (*tail*) terus menerus
make docker-down            # Menghentikan layanan *container*
make docker-remove-rebuild  # Membersihkan volume *database* dan meng-rebuild ulang dari nol
```

---

## Menjalankan secara Lokal (Tanpa Docker)

Bila Anda ingin mengembangkan aplikasi tanpa Docker, Anda memerlukan Go, Node.js, layanan server PostgreSQL yang berjalan, serta instalasi TeX Live pada mesin lokal Anda.

### 1. Instalasi TeX Live

```bash
# Ubuntu / Debian
make install-latex

# macOS
brew install --cask mactex
```

### 2. Menyiapkan Basis Data (*Database*)

Buat sebuah *database* PostgreSQL baru dan jalankan skrip struktural awalnya (`init.sql`):

```bash
psql -U postgres -c "CREATE DATABASE atstex_lab;"
psql -U postgres -d atstex_lab -f init.sql
```

### 3. Konfigurasi File `.env` Lokasi

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/atstex_lab?sslmode=disable
GOOGLE_CLIENT_ID=client-id-google-anda
GOOGLE_CLIENT_SECRET=rahasia-client-google-anda
GOOGLE_CALLBACK_URL=http://localhost:8080/auth/google/callback

# Konfigurasi Ekstraksi AI
AI_PROVIDER=openai
AI_MODEL=gpt-4o-mini
AI_API_KEY=api-key-anda
```

### 4. Menjalankan Server Publikasi Lokal

```bash
npm install        # instal Tailwind CSS (hanya diperulukan saat pertama kali)
make run           # Meng-compile CSS + menjalankan server pengembangan (*dev-server*) di port 8080
```

### 5. Membangun Berkas Aplikasi Binari yang Berdiri Sendiri (*Standalone Binary*)

```bash
make build
./atstex-lab       # Menjalankan langsung aplikasi binari hasil kompilasi (Go)
```

---

## Daftar Perintah Makefile

| Perintah | Deskripsi Fungsi |
|----------|------------------|
| `make run` | Melakukan *compile* CSS lalu menjalankan *development server* |
| `make build` | Melakukan *compile* CSS lalu melakukan *build* binari Go menghasilkan file eksekusi `./atstex-lab` |
| `make css` | Melakukan *compile* Tailwind CSS saja |
| `make tidy` | Menjalankan proses pembersihan *go mod tidy* |
| `make clean` | Menghapus arsip binari lama aplikasi yang disatukan |
| `make docker-run` | Memulai proses *Build image* dan menghidupkan eksekusi aplikasi dan instansi server PostgreSQL |
| `make docker-up` | Berfungsi sama dengan *docker-run* tetapi mengabaikan tahap pembuatan build (hanya menyalakan container dengan status image saat ini) |
| `make docker-down` | Menghentikan *containers* yang aktif |
| `make docker-remove-rebuild` | Menghapus semua kontainer serta datanya dan melakukan pemulihan build ulang. |
| `make docker-logs` | Menyajikan *live logs tail* |
| `make install-latex` | Skrip instalasi cepat khusus Linux Ubuntu/Debian server untuk sistem paket LaTeX TeX Live |

---

## Struktur Direktori Utama Proyek

```
atstex-lab/
├── cmd/server/main.go              # Titik masuk utama aplikasi (Main entry), rute API, dan middleware
├── internal/
│   ├── auth/auth.go                # Kendali Login, Logout, dan pengembalian (*Callback*) Google OAuth
│   ├── compiler/compiler.go        # Mesin pelaksana kompilasi LaTeX dalam batas folder sementara (*temp dir*)
│   ├── config/config.go            # Komponen sistem untuk pemuatan nilai konfigurasi *(file .env)*
│   ├── cvtemplate/                 # Rutinitas parser modul *loader* pemuat daftar berkas sistem CV
│   ├── domain/                     # Konstruksi relasional Data model utama (seperti *User, Session, CVProfile*)
│   ├── extractor/extractor.go      # Rutinitas Ekstraksi cerdas berbasis komputasi AI dari berkas teks PDF
│   ├── handler/                    # Konfigurasi rutinitas jalur pengendali HTTP / Web
│   └── repository/repository.go    # Pengendali sistem kueri pemrosesan transaksi *database* PostgreSQL
├── web/
│   ├── embed.go                    # Mekanisme internalisasi aset halaman dengan berkas Go Binary
│   ├── templates/                  # Induk susunan antarmuka *UI Web HTML* + direktori fail referensi tata-letak .tex
│   └── static/                     # Pusat aset statik aplikasi berupa desain CSS *(Tailwind)* dan integrasi fail JavaScript
├── init.sql                        # Cetak biru skema utama relasi susunan sistem kolom tabel database awal
├── Dockerfile                      # Spesifikasi instruksi ber-stage ganda (Multi-stage Go build → sistem operasi pengemasan final berisi berkas sistem kompilasi TeX Live)
├── compose.yml                     # Konfigurasi eksekusi perangkaian instansi wadah konektivitas wadah Docker
├── tailwind.config.js              # Penyesuaian kustom model gaya visual desain situs CSS
└── Makefile                        # Kendali modul alat pembantu otomatis fungsional perintah perintah bash shell
```

---

## Daftar Jalur *Endpoint* API Aplikasi

### Mode Halaman Antarmuka Publik (Pages)

| Jalur URL | Penjelasan Fungsi Endpoint |
|-----------|----------------------------|
| `GET /` | Merujuk ke muka sistem *(halaman beranda awal)* |
| `GET /input` | Memberikan akses ruang kontrol borang pengisian rekam formulir data profil CV utuh. |
| `GET /input/embed` | Format *UI Embed* formulir masukan pengisi identitas tanpa *sidebar*. |
| `GET /editor` | Tampilan muka untuk pengontrol *editor template LaTeX* dan pratinjau penyetelan penyusunan tampilan berkas formulir keluaran file output .PDF. |
| `GET /profile` | Daftar menu sistem pengaturan profil dasar klien pengguna. |

### Sistem Autentikasi Pengguna (Auth)

| Jalur URL | Penjelasan Fungsi Endpoint |
|-----------|----------------------------|
| `GET /auth/google/login` | Rute pemanggilan gerbang perizinan SSO Google Authentication. |
| `GET /auth/google/callback` | Jalur alur penerimaan sinyal kembalian (*callback redirect*) konfirmasi integrasi persetujuan pengguna pasca berhasil Login / Oauth flow. |
| `POST /auth/logout` | Titik perintah terminal keluar jaringan / penonaktifan sesi *login* di peramban saat ini. |
| `POST /auth/sessions/{token}/delete` | Perintah mutlak akses penghapusan memutus riwayat log koneksi integrasi di platform sistem eksternal pengguna. |

### API Profil CV Manajemen

| Jalur API | Penjelasan Fungsi Endpoint |
|-----------|----------------------------|
| `GET /api/cv-profiles` | Akses perolehan rincian senarai catatan koleksi profil CV milik pengguna. |
| `POST /api/cv-profiles` | Merintis pembuatan direktori profil rekam daftar entri baru (*payload:* `{"title": "..."}`). |
| `GET /api/cv-profiles/{id}` | Endpoint pengecekan informasi spesifik menyeluruh identitas profil yang dipanggil beserta biodata pendamping utuhnya. |
| `PUT /api/cv-profiles/{id}` | Rutinitas melakukan rekaman perubahan / pemutakhiran penyimpanan isian muatan informasi `{"biodata": {...}}`. |
| `DELETE /api/cv-profiles/{id}` | Penugasan permintaan mutlak pemusnahan keseluruhan direktori ID identitas *record* dari koleksi pengguna terotomatisasi. |

### API Template & Alat Kompilator (Compiler Tools)

| Jalur API | Penjelasan Fungsi Endpoint |
|-----------|----------------------------|
| `GET /api/templates` | Mendeteksi ketersediaan daftar parameter daftar silsilah tipe desain layout / kerangka form *template* yang bisa dipakai sistem. |
| `GET /api/templates/{name}` | Mengakses file inti *resource* pemrograman *script code markup LaTeX* bawaan standar secara langsung mentah-mentah dari nama varian yang diberikan. |
| `POST /api/templates/{name}/render` | Me-*render* modul file *blueprint template* lalu ditautkan (inject) dan dicetak format ke bentuk nilai susunan *payload text JSON biodata*. |
| `POST /api/extract-pdf` | Endpoint perutean integrasi konektivitas API kecerdasan buatan (*AI*) untuk menerjemahkan nilai pindaian arsip fisik dokumen teks tak berstruktur .pdf kemudian merekonstruksinya menjadi serangkaian parameter properti elemen nilai terarsik / *parsial payload* basis tata penguraian susunan model `biodata.json` dasar. |
| `POST /compile` | Menyuntik pemicu instansiasi (*trigger instantiation execution command node processor handler*) pemanggilan pengeksekutor sintaks perikator format markup dasar mesin program kode bahasa pemrograman komputer beridentifikator struktur *LaTeX resource document form template script syntax file*. Setelah dipanggil dipasangi perintah penanganan transmisi data komputasional *stream parsing converter response body rendering system handler application* untuk mempublikasikan kembali hasil perakitan final keluaran format instrumen dokumen sistem ke peninjau eksekusi klien sebagai tanggapan berupa salinan fisik cetak berwujud dokumen akhir visual file `.pdf` yang sudah diformat dan dikompilasi mulus siap unduh (*raw byte buffer document streaming print array transmission API*). |
