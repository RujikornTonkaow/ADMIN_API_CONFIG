# Tech Stack — Portfolio Admin API

เอกสารนี้อธิบาย Technology Stack ทั้งหมดที่โปรเจกต์นี้ใช้ รวมถึงเหตุผลในการเลือกใช้

---

## ภาพรวม Tech Stack

```
┌─────────────────────────────────────────────────────────────────┐
│                        Tech Stack Overview                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Language:       Go 1.22                                       │
│   HTTP Server:    net/http (standard library ServeMux)          │
│   Database:       MongoDB 7                                     │
│   Auth:           JWT (HS256) + bcrypt                          │
│   Containerize:   Docker + Docker Compose                       │
│   Logging:        log/slog (structured JSON)                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 1. Backend — Go 1.22

| หัวข้อ | รายละเอียด |
|-------|-----------|
| **ภาษา** | Go 1.22 |
| **HTTP Router** | `net/http` ServeMux พร้อม method-based routing (Go 1.22 feature) |
| **Pattern** | Clean Architecture: `handler → repository → database` |

### ทำไมถึงเลือก Go?
- Performance สูง — compiled language, low memory footprint
- Concurrency ดีเยี่ยม — goroutines สำหรับ concurrent requests
- Standard library มีครบ — `net/http` ทำ REST API ได้โดยไม่ต้องใช้ framework
- Go 1.22 ServeMux รองรับ method-based routing เช่น `"GET /api/v1/users/{id}"` ไม่ต้องพึ่ง third-party router
- Static binary — deploy ง่าย ไม่มี runtime dependency

### Go 1.22 ServeMux Features ที่ใช้
```go
// Method-based routing (ไม่ต้องเช็ค r.Method เอง)
mux.HandleFunc("GET /api/v1/public/sites/by-domain", handler.ResolveDomain)
mux.HandleFunc("GET /api/v1/public/sites/{siteId}/portfolio", handler.GetPortfolio)

// Path parameters
mux.HandleFunc("GET /api/v1/admin/users/{id}", handler.GetByID)
// เข้าถึงด้วย r.PathValue("id")
```

---

## 2. Database — MongoDB 7

| หัวข้อ | รายละเอียด |
|-------|-----------|
| **Version** | MongoDB 7 |
| **Go Driver** | `go.mongodb.org/mongo-driver` v1.17.9 |
| **Database Name** | `portfolio_admin` (configurable) |

### ทำไมถึงเลือก MongoDB?
- Schema-flexible — เหมาะกับ content management ที่โครงสร้างข้อมูลอาจเปลี่ยนบ่อย
- JSON-native — เก็บ/ดึงข้อมูลในรูปแบบ JSON/BSON ตรงกับ API response
- เหมาะกับ document-oriented data เช่น portfolio sections, projects, experiences

### Collections ที่ใช้
| Collection | ลักษณะ | คำอธิบาย |
|-----------|--------|---------|
| `admin_users` | Multiple docs | ข้อมูล admin users + unique index on username |
| `site_settings` | Singleton | ตั้งค่า site (title, theme, meta) |
| `hero` | Singleton | Hero section content |
| `about` | Singleton | About section content |
| `skills` | Multiple docs | Skills พร้อม sort_order |
| `projects` | Multiple docs | Projects พร้อม sort_order |
| `experiences` | Multiple docs | Work experiences พร้อม sort_order |
| `social_links` | Multiple docs | Social media links พร้อม sort_order |
| `contact_messages` | Multiple docs | ข้อความจาก contact form |

---

## 3. Authentication & Security

### JWT (JSON Web Token)

| หัวข้อ | รายละเอียด |
|-------|-----------|
| **Library** | `github.com/golang-jwt/jwt/v5` v5.3.1 |
| **Algorithm** | HMAC-SHA256 (HS256) |
| **Expiry** | 24 ชั่วโมง |
| **Secret** | `JWT_SECRET` env var |

**JWT Claims:**
```json
{
  "sub": "user_object_id",
  "usr": "username",
  "role": "super_admin|admin|editor|viewer",
  "exp": 1234567890,
  "iat": 1234567890
}
```

### bcrypt — Password Hashing

| หัวข้อ | รายละเอียด |
|-------|-----------|
| **Library** | `golang.org/x/crypto/bcrypt` v0.28.0 |
| **Cost** | `bcrypt.DefaultCost` (10 rounds) |
| **ใช้ตอน** | สร้าง user ใหม่, seed admin, login verification |

### RBAC (Role-Based Access Control)

| Role | Level | สิทธิ์ |
|------|-------|-------|
| `super_admin` | 4 | ทุกอย่าง ทุก site และสร้าง super_admin อื่นได้ |
| `admin` | 3 | จัดการ users/content เฉพาะ site ที่ได้รับสิทธิ์ |
| `editor` | 2 | แก้ content เฉพาะ site ที่ได้รับสิทธิ์ |
| `viewer` | 1 | ดู Messages/Contacts เฉพาะ site ที่ได้รับสิทธิ์ |

**การอนุญาตสองระดับ (Two-level authorization):**
- **Global RBAC** — บทบาทระดับบัญชีทั้งระบบใน JWT: `super_admin` / `admin` / `editor` / `viewer`
- **Site access** — `site_members` ระบุว่า user เข้าถึง site ไหนได้บ้าง

---

## 4. Unique ID Generation

| หัวข้อ | รายละเอียด |
|-------|-----------|
| **Library** | `github.com/google/uuid` v1.6.0 |
| **ใช้ตอน** | สร้าง Request ID (middleware), ตั้งชื่อไฟล์ upload |

---

## 5. Logging — Structured JSON

| หัวข้อ | รายละเอียด |
|-------|-----------|
| **Library** | `log/slog` (Go standard library) |
| **Format** | JSON output to stdout |
| **Level** | `slog.LevelInfo` |

**ข้อมูลที่ log ต่อ request:**
- `method` — HTTP method
- `path` — URL path
- `status` — HTTP status code
- `duration_ms` — เวลาที่ใช้ (milliseconds)
- `request_id` — UUID ของ request

**ตัวอย่าง log output:**
```json
{
  "time": "2025-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "request",
  "method": "GET",
  "path": "/api/v1/public/sites/{siteId}/portfolio",
  "status": 200,
  "duration_ms": 12,
  "request_id": "a1b2c3d4-e5f6-..."
}
```

---

## 6. Containerization — Docker

### Dockerfile (Multi-stage Build)

| Stage | Base Image | หน้าที่ |
|-------|-----------|---------|
| **Builder** | `golang:1.22-alpine` | Compile Go source → static binary |
| **Runtime** | `alpine:3.19` | รัน binary ขนาดเล็ก (~15MB) |

**Security:**
- Non-root user (`appuser`)
- Static binary (`CGO_ENABLED=0`)
- Minimal runtime image (Alpine)
- Healthcheck ทุก 30 วินาที

### Docker Compose

| Service | Image | Port | หน้าที่ |
|---------|-------|------|---------|
| `api` | Build จาก Dockerfile | `8080:8080` | Go API server |
| `mongo` | `mongo:7` | `27017:27017` | MongoDB database |

**Volumes:**
- `mongo_data` — persist ข้อมูล MongoDB
- `uploads` — persist ไฟล์ที่อัปโหลด

---

## 7. Go Dependencies Summary

```
go.mod dependencies:

github.com/golang-jwt/jwt/v5    v5.3.1    — JWT token creation & validation
github.com/google/uuid           v1.6.0    — UUID generation
go.mongodb.org/mongo-driver     v1.17.9   — MongoDB official Go driver
golang.org/x/crypto             v0.28.0   — bcrypt password hashing
```

---

## 8. Development Tools ที่แนะนำ

| Tool | หน้าที่ | ดาวน์โหลด |
|------|---------|----------|
| **Go 1.22+** | ภาษาหลักของ backend | [go.dev/dl](https://go.dev/dl/) |
| **Docker Desktop** | รัน containers (API + MongoDB) | [docker.com](https://www.docker.com/products/docker-desktop/) |
| **MongoDB Compass** | GUI สำหรับดู/จัดการ MongoDB data | [mongodb.com/compass](https://www.mongodb.com/products/compass) |
| **Postman / Bruno** | ทดสอบ API endpoints | [postman.com](https://www.postman.com/) |
| **VS Code / Cursor** | Code editor | [code.visualstudio.com](https://code.visualstudio.com/) |
| **Git** | Version control | [git-scm.com](https://git-scm.com/) |

### VS Code / Cursor Extensions ที่แนะนำ
- **Go** (`golang.go`) — Go language support, IntelliSense, debugging
- **REST Client** — ทดสอบ API จาก editor
- **Docker** — จัดการ Docker containers
- **MongoDB for VS Code** — เชื่อมต่อ MongoDB จาก editor

---

## 9. Architecture Patterns ที่ใช้

| Pattern | ที่ใช้ | คำอธิบาย |
|---------|-------|---------|
| **Clean Architecture** | ทั้งโปรเจกต์ | แบ่ง layer: Handler → Repository → Database |
| **Repository Pattern** | `internal/repository/` | แยก data access logic ออกจาก business logic |
| **Middleware Pattern** | `internal/middleware/` | chain middleware สำหรับ cross-cutting concerns |
| **Multi-Site / Multi-Tenant** | ทั้งระบบ | แยกข้อมูลต่อไซต์ด้วยฟิลด์ `site_id`; ตรวจสอบสมาชิกไซต์ผ่าน middleware |
| **Dependency Injection** | `cmd/server/main.go` | inject dependencies ผ่าน constructor functions |
| **Singleton Pattern** | site_settings, hero, about | collection ที่มี document เดียว ใช้ upsert |
| **Envelope Pattern** | `pkg/response/` | JSON response ครอบด้วย `{ data, error, meta }` |

**Middleware ที่เกี่ยวกับไซต์:** `RequireSiteMember` — ตรวจสอบว่า user มี access record ใน `site_members` สำหรับ site ที่ request อ้างถึง โดย `super_admin` bypass ได้ ส่วนความสามารถอ่าน/เขียนตัดสินจาก global role (`admin`, `editor`, `viewer`)

---

## 10. Frontend ที่คาดว่าจะเชื่อมต่อ

โปรเจกต์นี้เป็น **Backend API เท่านั้น** ไม่มี frontend รวมอยู่ด้วย

CORS ตั้งค่า (`ALLOWED_ORIGINS` — static origins หลายค่า คั่นด้วย comma):
- **ตัวอย่างค่าแนะนำ:** `http://localhost:3001,https://admin.example.com`
- **Go config default:** `http://localhost:3000` (ถ้าไม่ตั้ง env ใน `config.go`)
- **Portfolio domains:** ไม่ต้องเพิ่มทุกโดเมนใน env ถ้าเป็น site ในระบบ เพราะ `DomainCache` โหลดจาก `sites.domains`
- `http://localhost:3001` — Admin dashboard (SPA) local
- `https://admin.example.com` — ตัวอย่าง admin บน production

**หมายเหตุ:** ถ้าไม่มี `.env` ระบบจะใช้ค่า default จาก `config.go` คือ `http://localhost:3000` เท่านั้น — production ควรตั้ง `ALLOWED_ORIGINS` ให้ครอบคลุม admin/static origins และจัดการ portfolio domains ผ่าน `Site Management`

**Tech stack ที่แนะนำสำหรับ frontend:**
- **Nuxt 3** + **Vue 3** + **TypeScript** + **TailwindCSS**
- ใช้ `useFetch()` เรียก API
- เก็บ JWT token ใน cookie หรือ localStorage
- ใช้ middleware ตรวจสอบ auth ก่อนเข้าหน้า admin
