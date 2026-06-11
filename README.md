# Task Tracker API — Backend (Golang)

REST API untuk Task Tracker App, dibangun sebagai bagian dari Technical Test Fullstack Developer (Flutter focus).

---

## Cara Menjalankan Project

### Prasyarat

- [Go 1.21+](https://go.dev/dl/) — untuk menjalankan secara lokal
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) — untuk Docker Compose

### Option A — Docker Compose (Recommended)

Menjalankan PostgreSQL + API sekaligus tanpa perlu install Go atau setup database manual.

```bash
docker compose up --build
```

API berjalan di `http://localhost:8080/api/v1`

```bash
# Jalankan di background
docker compose up --build -d

# Stop
docker compose down

# Stop + hapus data volume
docker compose down -v
```

### Option B — Local Go (In-Memory, tanpa database)

Tidak butuh database. Data di-seed otomatis dengan 3 task contoh. Cocok untuk development cepat.

```bash
go mod tidy
go run main.go
```

> ⚠️ Data hilang saat server di-restart.

### Option C — Local Go + PostgreSQL

```bash
# 1. Jalankan PostgreSQL via Docker
docker run -d --name task-pg \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=tasktracker \
  -p 5433:5432 \
  postgres:16-alpine

# 2. Jalankan API dengan mode PostgreSQL
USE_POSTGRES=true DB_PORT=5433 go run main.go
```

### Menjalankan Tests

```bash
# Semua tests
go test ./...

# Verbose
go test ./... -v

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Konfigurasi Environment Variables

| Variable | Default | Keterangan |
|---|---|---|
| `USE_POSTGRES` | *(kosong)* | Set ke nilai apapun untuk pakai PostgreSQL |
| `DB_HOST` | `localhost` | Hostname PostgreSQL |
| `DB_PORT` | `5432` | Port PostgreSQL |
| `DB_USER` | `postgres` | Username database |
| `DB_PASSWORD` | `postgres` | Password database |
| `DB_NAME` | `tasktracker` | Nama database |

---

## Architecture Explanation

Project menggunakan **layered architecture** dengan separation yang jelas antar layer:

```
backend/
├── model/                            # Domain layer
│   └── task.go                       # Struct Task, request/response, validasi
├── repository/                       # Data layer
│   ├── task_repository.go            # Implementasi in-memory
│   ├── postgres_task_repository.go   # Implementasi PostgreSQL
│   └── task_repository_test.go
├── handler/                          # Presentation layer
│   ├── task_handler.go               # HTTP handler + interface TaskRepo
│   └── task_handler_test.go
├── db/                               # Infrastruktur
│   └── migrate.go                    # Koneksi pool + DDL migration
├── docs/
│   └── openapi.yaml                  # Spesifikasi OpenAPI 3.0
└── main.go                           # Entry point — wiring semua layer
```

### Alur Request

```
HTTP Request
    │
    ▼
Handler (parse, validasi, format response)
    │  memanggil via interface TaskRepo
    ▼
Repository (akses storage)
    │  in-memory atau PostgreSQL tergantung env
    ▼
Model (struct data, aturan validasi)
    │
    ▼
HTTP Response
```

### Interface-Based Design

Handler tidak bergantung ke implementasi konkret, melainkan ke interface:

```go
type TaskRepo interface {
    GetAll() []model.Task
    GetByID(id string) (model.Task, error)
    Create(req model.CreateTaskRequest) model.Task
    UpdateStatus(id string, status model.TaskStatus) (model.Task, error)
    Delete(id string) error
}
```

Keuntungan: storage bisa diganti tanpa ubah handler, dan handler test bisa menggunakan mock tanpa database nyata.

---

## State Management Explanation

Backend tidak memiliki state management dalam artian yang sama dengan Flutter. Tidak ada session, tidak ada in-memory application state yang di-share antar request.

Setiap HTTP request diproses secara **independen dan stateless** — handler menerima request, memanggil repository, lalu mengembalikan response. Tidak ada data yang disimpan di level aplikasi antara satu request dengan request berikutnya.

**Thread-safety** untuk in-memory repository dikelola dengan `sync.RWMutex` — bukan sebagai state management, tapi sebagai mekanisme untuk mencegah race condition saat multiple goroutine mengakses slice yang sama secara bersamaan.

> State management yang dimaksud dalam requirement teknikal test ini ada di sisi **Flutter** — lihat [frontend README](../frontend/task_tracker_app/README.md) untuk penjelasan lengkapnya.

---

## Alasan Memilih Approach Tertentu

### Gin sebagai HTTP Framework
Dipilih karena ringan, performant, dan idiomatik untuk REST API di Go. Middleware chain-nya clean, dan built-in validation via struct binding tag (`required`, `min`, `max`, `oneof`) mengurangi boilerplate validasi.

### Dual Storage Strategy
Toggle via env var `USE_POSTGRES`. Alasannya:
- Developer bisa langsung `go run main.go` tanpa setup apapun
- In-memory sudah di-seed dengan data contoh sehingga frontend bisa langsung integrasi
- Production menggunakan PostgreSQL yang data-nya persisten
- Unit test berjalan cepat tanpa koneksi database

### pgxpool untuk PostgreSQL
Gin melayani HTTP request secara concurrent. Single connection menjadi bottleneck; pool koneksi memungkinkan multiple query berjalan paralel secara efisien.

### RETURNING pada Query UPDATE dan INSERT
Query `UPDATE ... RETURNING` dan `INSERT ... RETURNING` mengembalikan data yang sudah tersimpan dalam satu roundtrip ke database, tanpa perlu SELECT terpisah setelahnya. Lebih efisien dan atomic.

### Retry Connect ke PostgreSQL
Saat menggunakan Docker Compose, container backend sering start lebih cepat dari PostgreSQL. Fungsi `Connect()` melakukan retry hingga 10 kali dengan interval 2 detik — menghindari crash saat startup tanpa perlu `sleep` di entrypoint script.

---

## Endpoint API

Base URL: `http://localhost:8080/api/v1`

| Method | Endpoint | Deskripsi | Response |
|---|---|---|---|
| GET | `/tasks` | List semua task (newest first) | 200 |
| GET | `/tasks/:id` | Detail task | 200, 404 |
| POST | `/tasks` | Buat task baru | 201, 400 |
| PATCH | `/tasks/:id/status` | Update status task | 200, 400, 404 |
| DELETE | `/tasks/:id` | Hapus task | 200, 404 |
| GET | `/health` | Health check | 200 |

### Contoh Request

```bash
# List semua task
curl http://localhost:8080/api/v1/tasks

# Buat task baru
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "New Task", "description": "Task description"}'

# Update status
curl -X PATCH http://localhost:8080/api/v1/tasks/{id}/status \
  -H "Content-Type: application/json" \
  -d '{"status": "done"}'
```

Dokumentasi lengkap tersedia di [`docs/openapi.yaml`](docs/openapi.yaml).

---

## Tests

| Package | Test Cases | Total |
|---|---|---|
| `handler` | GetAll (2), GetByID (2), Create (4), UpdateStatus (3), Delete (2) | 13 |
| `repository` | GetAll (2), GetByID (2), Create (2), UpdateStatus (2), Delete (3) | 11 |
| **Total** | | **23 tests** |

Handler test menggunakan **mock repository** yang mengimplementasikan interface `TaskRepo` — tidak membutuhkan database atau server nyata.

---

## Teknologi

| Komponen | Versi |
|---|---|
| Go | 1.21+ |
| Gin | 1.9.1 |
| gin-contrib/cors | 1.5.0 |
| google/uuid | 1.6.0 |
| pgx/v5 | 5.6.0 |
| PostgreSQL | 16 |
