# File Reference — Portfolio Admin API

เอกสารนี้อธิบายรายละเอียดของทุกไฟล์ในโปรเจกต์ว่าแต่ละไฟล์ทำหน้าที่อะไร

---

## โครงสร้าง Directory ทั้งหมด

```
portfolio-admin-api/
├── .env.example                 # ตัวอย่าง environment variables
├── .gitignore                   # กฎสำหรับไฟล์ที่ไม่ต้อง track ใน git
├── Dockerfile                   # สร้าง Docker image (multi-stage build)
├── docker-compose.yml           # รันทั้ง API + MongoDB ด้วย Docker
├── go.mod                       # Go module dependencies
├── go.sum                       # Checksum ของ dependencies (auto-generated)
├── README.md                    # เอกสารหลักของโปรเจกต์
├── RBAC_MIGRATION_GUIDE.md      # คู่มือ RBAC สำหรับ frontend
│
├── cmd/
│   └── server/
│       └── main.go              # ★ Entry point ของ application
│
├── internal/
│   ├── config/
│   │   └── config.go            # โหลด configuration จาก env vars
│   │
│   ├── database/
│   │   ├── mongodb.go           # เชื่อมต่อ MongoDB
│   │   └── seed.go              # Seed ข้อมูลเริ่มต้น
│   │
│   ├── handler/
│   │   ├── about.go             # จัดการ About section
│   │   ├── auth.go              # Login + ดูข้อมูล user ปัจจุบัน
│   │   ├── contact.go           # จัดการ Contact messages
│   │   ├── experience.go        # จัดการ Experience (CRUD + reorder)
│   │   ├── hero.go              # จัดการ Hero section
│   │   ├── project.go           # จัดการ Projects (CRUD + reorder)
│   │   ├── public.go            # API สาธารณะ (portfolio + contact form)
│   │   ├── site.go              # จัดการ Sites (CRUD)
│   │   ├── site_member.go       # จัดการ Site Members
│   │   ├── site_settings.go     # จัดการ Site Settings
│   │   ├── skill.go             # จัดการ Skills (CRUD)
│   │   ├── social_link.go       # จัดการ Social Links (CRUD + reorder)
│   │   ├── upload.go            # อัปโหลดไฟล์ภาพ
│   │   └── user.go              # จัดการ Admin Users (CRUD)
│   │
│   ├── middleware/
│   │   └── middleware.go        # Middleware ทั้งหมด (CORS, Auth, RBAC ฯลฯ)
│   │
│   ├── model/
│   │   └── model.go             # Data models + DTOs + Role constants
│   │
│   ├── repository/
│   │   ├── about.go             # MongoDB operations สำหรับ about
│   │   ├── admin_user.go        # MongoDB operations สำหรับ admin_users
│   │   ├── contact.go           # MongoDB operations สำหรับ contact_messages
│   │   ├── experience.go        # MongoDB operations สำหรับ experiences
│   │   ├── hero.go              # MongoDB operations สำหรับ hero
│   │   ├── project.go           # MongoDB operations สำหรับ projects
│   │   ├── site.go              # MongoDB operations สำหรับ sites
│   │   ├── site_member.go       # MongoDB operations สำหรับ site_members
│   │   ├── site_settings.go     # MongoDB operations สำหรับ site_settings
│   │   ├── skill.go             # MongoDB operations สำหรับ skills
│   │   └── social_link.go       # MongoDB operations สำหรับ social_links
│   │
│   └── router/
│       └── router.go            # Route registration + middleware wiring
│
├── pkg/
│   └── response/
│       └── json.go              # JSON response helpers (envelope format)
│
├── uploads/                     # เก็บไฟล์ที่อัปโหลด (gitignored)
│
└── GUIDE/                       # เอกสาร guide สำหรับนักพัฒนา
```

---

## รายละเอียดแต่ละไฟล์

### Root Files

| ไฟล์ | หน้าที่ |
|------|---------|
| `.env.example` | ตัวอย่าง environment variables ทั้งหมดที่ระบบต้องการ ให้ copy เป็น `.env` แล้วแก้ค่า |
| `.gitignore` | กำหนดไฟล์ที่ไม่ต้อง commit เช่น `.env`, `uploads/*`, binary files, IDE configs |
| `Dockerfile` | Multi-stage build: build ด้วย `golang:1.22-alpine` → runtime ด้วย `alpine:3.19` พร้อม healthcheck |
| `docker-compose.yml` | กำหนด 2 services: `api` (Go app) + `mongo` (MongoDB 7) พร้อม volumes และ healthcheck |
| `go.mod` | ประกาศ module name (`portfolio-admin-api`) และ dependencies ที่ใช้ |
| `go.sum` | Auto-generated checksum file สำหรับ verify dependencies |
| `README.md` | เอกสารภาพรวมโปรเจกต์, วิธีรัน, API overview |
| `RBAC_MIGRATION_GUIDE.md` | คู่มือสำหรับ frontend เกี่ยวกับระบบ RBAC และ user management API |

---

### `cmd/server/main.go` — Application Entry Point

**หน้าที่:** เป็นจุดเริ่มต้นของแอปพลิเคชัน ทำหน้าที่ orchestrate ทุกอย่าง

**สิ่งที่ทำ (ตามลำดับ):**
1. สร้าง structured JSON logger
2. โหลด config จาก environment variables
3. สร้าง upload directory
4. เชื่อมต่อ MongoDB
5. สร้าง indexes บน collections `admin_users`, `sites`, `site_members`
6. Seed ข้อมูลเริ่มต้น (ถ้า database ว่าง)
7. สร้าง repositories ทั้งหมด
8. สร้าง router พร้อม middleware
9. เริ่ม HTTP server
10. รอ shutdown signal แล้วปิดอย่าง graceful

---

### `internal/config/config.go` — Configuration Loader

**หน้าที่:** โหลดค่า configuration ทั้งหมดจาก environment variables พร้อมค่า default

**Config fields:**

| Field | Env Var | Default | คำอธิบาย |
|-------|---------|---------|---------|
| `Port` | `PORT` | `8080` | พอร์ตที่ server listen |
| `MongoURI` | `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `MongoDB` | `MONGO_DB` | `portfolio_admin` | ชื่อ database |
| `JWTSecret` | `JWT_SECRET` | `change-me-in-production` | Secret key สำหรับ sign JWT |
| `AdminUsername` | `ADMIN_USERNAME` | `admin` | Username สำหรับ seed admin |
| `AdminPassword` | `ADMIN_PASSWORD` | `changeme123` | Password สำหรับ seed admin |
| `UploadDir` | `UPLOAD_DIR` | `./uploads` | Directory สำหรับเก็บไฟล์อัปโหลด |
| `MaxUploadSizeMB` | `MAX_UPLOAD_SIZE_MB` | `10` | ขนาดไฟล์สูงสุดที่อัปโหลดได้ (MB) |
| `AllowedOrigins` | `ALLOWED_ORIGINS` | `http://localhost:3000` | CORS allowed origins (คั่นด้วย comma) |

---

### `internal/database/mongodb.go` — Database Connection

**หน้าที่:** จัดการการเชื่อมต่อกับ MongoDB

**สิ่งที่ทำ:**
- `Connect()` — เชื่อมต่อ MongoDB ด้วย timeout 10 วินาที, ping เพื่อตรวจสอบ, return database instance + disconnect function

---

### `internal/database/seed.go` — Database Seeder

**หน้าที่:** สร้างข้อมูลเริ่มต้นเมื่อ database ว่าง (first run)

**สิ่งที่ seed:**
- `admin_users` — สร้าง admin user (bcrypt hashed password, role: admin)
- `site_settings` — ค่า default ของ site (title, theme, meta description)
- `hero` — ข้อมูล hero section (greeting, name, subtitle, CTA buttons)
- `about` — bio paragraphs, personality tags, stats
- `skills` — 14 skills ตัวอย่าง (Vue.js, Go, Docker ฯลฯ) แบ่งตาม category
- `projects` — 2 projects ตัวอย่าง
- `experiences` — 2 experiences ตัวอย่าง
- `social_links` — 4 links ตัวอย่าง (GitHub, LinkedIn, Twitter, Email)

**เงื่อนไข:** จะ seed เฉพาะเมื่อ `admin_users` collection มี 0 documents เท่านั้น

---

### `internal/handler/` — Request Handlers (Business Logic)

ทุก handler ใน folder นี้ทำหน้าที่:
1. รับ HTTP request
2. Validate input
3. เรียก repository layer
4. ส่ง JSON response

| ไฟล์ | Endpoints | คำอธิบาย |
|------|-----------|---------|
| `public.go` | `GET /public/sites/{siteId}/portfolio`, `POST /public/sites/{siteId}/portfolio/contacts`, `GET /public/sites/by-domain` | API สาธารณะ: ดึงข้อมูล portfolio ตาม site + ส่ง contact message + ค้นหา site จาก domain |
| `auth.go` | `POST /login`, `GET /me` | Login ด้วย username/password → JWT token, ดูข้อมูล user ปัจจุบัน |
| `user.go` | CRUD `/admin/users` | จัดการ admin users (สร้าง, ดู, แก้ไข, ลบ, เปลี่ยนรหัสผ่าน) — admin only |
| `site.go` | CRUD `/admin/sites` | จัดการ sites (สร้าง, ดู, แก้ไข, ลบ) |
| `site_member.go` | CRUD `/admin/sites/{siteId}/members` | จัดการสมาชิกของ site (เพิ่ม, ดู, แก้ role, ลบ) |
| `site_settings.go` | GET/PUT `/admin/sites/{siteId}/portfolio/site-settings` | ดู/แก้ไข site settings (site title, theme, meta) |
| `hero.go` | GET/PUT `/admin/sites/{siteId}/portfolio/hero` | ดู/แก้ไข hero section (greeting, name, CTA) |
| `about.go` | GET/PUT `/admin/sites/{siteId}/portfolio/about` | ดู/แก้ไข about section (bio, tags, stats) |
| `skill.go` | CRUD `/admin/sites/{siteId}/portfolio/skills` | จัดการ skills (สร้าง, ดู, แก้ไข, ลบ) |
| `project.go` | CRUD+Reorder `/admin/sites/{siteId}/portfolio/projects` | จัดการ projects (สร้าง, ดู, แก้ไข, ลบ, เรียงลำดับ) |
| `experience.go` | CRUD+Reorder `/admin/sites/{siteId}/portfolio/experiences` | จัดการ experiences (สร้าง, ดู, แก้ไข, ลบ, เรียงลำดับ) |
| `social_link.go` | CRUD+Reorder `/admin/sites/{siteId}/portfolio/social-links` | จัดการ social links (สร้าง, ดู, แก้ไข, ลบ, เรียงลำดับ) |
| `contact.go` | List/Get/Delete `/admin/sites/{siteId}/portfolio/contacts` | ดูรายการ/รายละเอียด/ลบ contact messages |
| `upload.go` | `POST /admin/sites/{siteId}/portfolio/upload` | อัปโหลดรูปภาพ: validate type/size → UUID filename → save to disk |

---

### `internal/middleware/middleware.go` — Middleware Stack

**หน้าที่:** ฟังก์ชัน middleware ทั้งหมดที่ครอบ HTTP handlers

| Middleware | หน้าที่ |
|-----------|---------|
| `Chain()` | รวม middleware หลายตัวเป็น chain เรียงจากนอกเข้าใน |
| `RequestID()` | สร้าง UUID v4 ใส่ context และ `X-Request-ID` header |
| `Logging()` | Log ทุก request: method, path, status, duration_ms, request_id |
| `Recovery()` | ดักจับ panic ป้องกัน server crash, log error + ส่ง 500 |
| `CORS()` | ตรวจสอบ Origin, set Access-Control headers, handle preflight OPTIONS |
| `Auth()` | ตรวจสอบ JWT token จาก Authorization header, inject user info ลง context |
| `RequireRole()` | ตรวจสอบ role level ของ user ว่าเพียงพอสำหรับ endpoint นั้น |
| `RequireSiteMember()` | ตรวจสอบสิทธิ์ระดับ site (owner >= editor >= viewer) + inject `siteID` ลง context |

**Helper functions:**
- `GetRequestID(ctx)` — ดึง request ID จาก context
- `GetUserID(ctx)` — ดึง user ID จาก context
- `GetUsername(ctx)` — ดึง username จาก context
- `GetRole(ctx)` — ดึง role จาก context
- `GetSiteID(ctx)` — ดึง site ID จาก context

---

### `internal/model/model.go` — Data Models & DTOs

**หน้าที่:** กำหนดโครงสร้างข้อมูลทั้งหมดของระบบ

**Multi-site Models:**
- `Site` — ข้อมูล site (เช่น slug, domain, metadata)
- `SiteMember` — ความสัมพันธ์ user กับ site และ role ระดับ site

**Singleton Models (เอกสารเดียวต่อ collection ต่อ site):**
- `SiteSettings` — ชื่อไซต์, page title, meta description, theme, profile image
- `Hero` — greeting, ชื่อ, subtitle, CTA buttons
- `About` — title, bio paragraphs, personality tags, stats

**Collection Models (หลายเอกสารต่อ site):**
- `Skill` — name, icon, category, sort_order
- `Project` — title, description, tags, image, URLs, sort_order
- `Experience` — role, company, period, description, highlights, sort_order
- `SocialLink` — name, URL, icon, sort_order
- `ContactMessage` — name, email, subject, message, is_read
- `AdminUser` — username, password (hashed), role

**หมายเหตุ:** โมเดล portfolio ที่มีอยู่เดิม (`SiteSettings`, `Hero`, `About`, `Skill`, `Project`, `Experience`, `SocialLink`, `ContactMessage`) มีฟิลด์ `SiteID` เพื่อผูกข้อมูลกับ site

**Site role constants:**
- `SiteRoleOwner`, `SiteRoleEditor`, `SiteRoleViewer` (legacy storage compatibility; UI ใช้ site access list)

**RBAC Constants:**
- `RoleSuperAdmin = "super_admin"` (Level 4)
- `RoleAdmin = "admin"` (Level 3)
- `RoleEditor = "editor"` (Level 2)
- `RoleViewer = "viewer"` (Level 1)

**API DTOs (Request/Response):**
- `LoginRequest`, `LoginResponse`, `LoginUser`
- `CreateUserRequest`, `UpdateUserRequest`, `ChangePasswordRequest`
- `ReorderRequest`, `ContactRequest`
- `PortfolioData`, `NavItem`

---

### `internal/repository/` — Data Access Layer

ทุกไฟล์ใน folder นี้ทำหน้าที่เป็น **Data Access Object (DAO)** สำหรับ MongoDB collection ที่เกี่ยวข้อง

| ไฟล์ | Collection | Operations |
|------|-----------|-----------|
| `about.go` | `about` | Get (FindOne), Upsert (update หรือ insert) — รับ `siteID` |
| `admin_user.go` | `admin_users` | EnsureIndexes, Create, FindByUsername, FindByID, List, Update, UpdatePassword, Delete, CountByRole |
| `contact.go` | `contact_messages` | Create, List, GetByID, MarkAsRead, Delete, CountUnread — รับ `siteID` |
| `experience.go` | `experiences` | Create, List, FindByID, Update, Delete, Reorder — รับ `siteID` |
| `hero.go` | `hero` | Get (FindOne), Upsert (update หรือ insert) — รับ `siteID` |
| `project.go` | `projects` | Create, List, FindByID, Update, Delete, Reorder — รับ `siteID` |
| `site.go` | `sites` | EnsureIndexes, Create, FindByID, FindBySlug, FindByDomain, ListByIDs, Update, Delete |
| `site_member.go` | `site_members` | EnsureIndexes, Create, FindByID, FindBySiteAndUser, ListBySite, ListByUser, UpdateRole, Delete, DeleteBySite, CountOwners |
| `site_settings.go` | `site_settings` | Get (FindOne), Upsert (update หรือ insert) — รับ `siteID` |
| `skill.go` | `skills` | Create, List, FindByID, Update, Delete — รับ `siteID` |
| `social_link.go` | `social_links` | Create, List, FindByID, Update, Delete, Reorder — รับ `siteID` |

**หมายเหตุ:** repository ของ portfolio ที่มีอยู่เดิมรับพารามิเตอร์ `siteID` เพื่อกรอง/เขียนข้อมูลตาม site

**หมายเหตุ:** Singleton collections (about, hero, site_settings) ใช้ pattern แยก:
- `Get` = FindOne จาก collection (ถ้าไม่มี document จะ error)
- `Upsert` = Update ถ้ามีอยู่แล้ว หรือ Insert ถ้ายังไม่มี
- ข้อมูลเริ่มต้นถูกสร้างโดย seed function ตอน first run

---

### `internal/router/router.go` — Route Registration

**หน้าที่:** ลงทะเบียน routes ทั้งหมด, ตั้งค่า RBAC middleware, สร้าง handler instances, mount static file server

**สิ่งที่ทำ:**
1. สร้าง auth middleware (JWT validation)
2. สร้าง role-based middleware: `requireAdmin`, `requireUser`, `requireVisitor`
3. สร้าง site-scoped middleware: `siteOwner`, `siteEditor`, `siteViewer` (ใช้ร่วมกับ `RequireSiteMember()` ตามระดับสิทธิ์)
4. สร้าง handler instances ทั้งหมดพร้อม inject dependencies
5. ลงทะเบียน routes ตาม HTTP method + path — public และ admin ที่เกี่ยวกับ portfolio เป็นแบบ site-scoped (`/public/sites/...`, `/admin/sites/{siteId}/portfolio/...`)
6. Mount static file server สำหรับ `/uploads/`
7. ครอบ global middleware stack: Recovery → CORS → Logging → RequestID

---

### `pkg/response/json.go` — JSON Response Helpers

**หน้าที่:** Helper functions สำหรับส่ง JSON response ในรูปแบบ Envelope

| Function | คำอธิบาย |
|----------|---------|
| `JSON(w, status, data)` | ส่ง success response: `{ "data": ... }` |
| `Error(w, status, msg)` | ส่ง error response: `{ "error": "..." }` |
| `JSONWithMeta(w, status, data, meta)` | ส่ง response พร้อม metadata: `{ "data": ..., "meta": ... }` |
| `DecodeJSON(r, dst)` | อ่าน JSON body จาก request (DisallowUnknownFields) |

---

### Docker & Deployment Files

**`Dockerfile`:**
- **Build stage:** ใช้ `golang:1.22-alpine`, compile เป็น static binary (`CGO_ENABLED=0`)
- **Runtime stage:** ใช้ `alpine:3.19`, สร้าง non-root user (`appuser`), expose port 8080
- **Healthcheck:** `wget` ไปที่ `/api/v1/portfolio` ทุก 30 วินาที

**`docker-compose.yml`:**
- **api service:** build จาก Dockerfile, port 8080, environment variables, uploads volume
- **mongo service:** `mongo:7`, port 27017, data volume, healthcheck ด้วย `mongosh ping`
- **Volumes:** `mongo_data` (เก็บข้อมูล DB), `uploads` (เก็บไฟล์อัปโหลด)
