# Portfolio Admin API

Backend API สำหรับจัดการเนื้อหาเว็บ Portfolio ผ่าน Admin Panel (CMS)

## Tech Stack

| Technology | Purpose |
|-----------|---------|
| Go 1.22+ | API Server (net/http with Go 1.22 ServeMux) |
| MongoDB 7 | Database |
| JWT | Authentication |
| Docker | Containerization |

## Quick Start

### ด้วย Docker Compose (แนะนำ)

```bash
# สร้าง .env จาก .env.example
cp .env.example .env

# แก้ไข JWT_SECRET และ ADMIN_PASSWORD ใน .env

# รัน
docker compose up -d
```

API จะพร้อมใช้งานที่ `http://localhost:8080`

### ด้วย Go (Development)

ต้องการ Go 1.22+ และ MongoDB ที่รันอยู่

```bash
# ติดตั้ง dependencies
go mod tidy

# รัน
go run ./cmd/server
```

## API Endpoints

### Public (ไม่ต้อง auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/portfolio` | ดึงข้อมูล portfolio ทั้งหมด |
| POST | `/api/v1/contact` | ส่งข้อความจาก contact form |

### Auth

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/admin/auth/login` | เข้าสู่ระบบ (ได้ JWT token) |

### Admin — Site Settings (ต้อง auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/site-settings` | ดึง site settings |
| PUT | `/api/v1/admin/site-settings` | อัพเดท site settings |

### Admin — Hero (ต้อง auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/hero` | ดึง hero data |
| PUT | `/api/v1/admin/hero` | อัพเดท hero data |

### Admin — About (ต้อง auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/about` | ดึง about data |
| PUT | `/api/v1/admin/about` | อัพเดท about data |

### Admin — Skills CRUD (ต้อง auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/skills` | ดึง skills ทั้งหมด |
| POST | `/api/v1/admin/skills` | เพิ่ม skill |
| PUT | `/api/v1/admin/skills/{id}` | แก้ไข skill |
| DELETE | `/api/v1/admin/skills/{id}` | ลบ skill |

### Admin — Projects CRUD (ต้อง auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/projects` | ดึง projects ทั้งหมด |
| POST | `/api/v1/admin/projects` | เพิ่ม project |
| PUT | `/api/v1/admin/projects/{id}` | แก้ไข project |
| DELETE | `/api/v1/admin/projects/{id}` | ลบ project |
| PUT | `/api/v1/admin/projects/reorder` | จัดลำดับ projects |

### Admin — Experiences CRUD (ต้อง auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/experiences` | ดึง experiences ทั้งหมด |
| POST | `/api/v1/admin/experiences` | เพิ่ม experience |
| PUT | `/api/v1/admin/experiences/{id}` | แก้ไข experience |
| DELETE | `/api/v1/admin/experiences/{id}` | ลบ experience |
| PUT | `/api/v1/admin/experiences/reorder` | จัดลำดับ experiences |

### Admin — Social Links CRUD (ต้อง auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/social-links` | ดึง social links ทั้งหมด |
| POST | `/api/v1/admin/social-links` | เพิ่ม social link |
| PUT | `/api/v1/admin/social-links/{id}` | แก้ไข social link |
| DELETE | `/api/v1/admin/social-links/{id}` | ลบ social link |

### Admin — Contact Messages (ต้อง auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/contacts` | ดึงข้อความทั้งหมด |
| GET | `/api/v1/admin/contacts/{id}` | ดูข้อความ (auto mark read) |
| DELETE | `/api/v1/admin/contacts/{id}` | ลบข้อความ |

### Admin — File Upload (ต้อง auth)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/admin/upload` | อัพโหลดรูปภาพ |
| GET | `/uploads/{filename}` | ดูรูปที่อัพโหลด (public) |

## Authentication

1. Login เพื่อรับ token:

```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"changeme123"}'
```

2. ใช้ token ใน Header:

```
Authorization: Bearer <token>
```

## Response Format

ทุก response ใช้ JSON envelope format:

```json
{
  "data": { ... },
  "error": "error message (ถ้ามี)",
  "meta": { ... }
}
```

## Project Structure

```
├── cmd/server/          # Entry point
├── internal/
│   ├── config/          # Environment configuration
│   ├── database/        # MongoDB connection + seed data
│   ├── handler/         # HTTP handlers (request/response)
│   ├── middleware/       # HTTP middleware stack
│   ├── model/           # Domain models
│   ├── repository/      # Data access layer (MongoDB)
│   └── router/          # Route registration
├── pkg/response/        # JSON response helpers
├── uploads/             # Uploaded files directory
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose with MongoDB
└── .env.example         # Environment variables template
```

## Seeding

เมื่อรัน API ครั้งแรก ระบบจะ seed ข้อมูลเริ่มต้นอัตโนมัติ:
- Admin user (จาก env vars `ADMIN_USERNAME` / `ADMIN_PASSWORD`)
- Site settings, Hero, About data
- Skills 14 รายการ
- Projects 2 รายการ
- Experiences 2 รายการ
- Social links 4 รายการ

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection URI |
| `MONGO_DB` | `portfolio_admin` | Database name |
| `JWT_SECRET` | `change-me-in-production` | JWT signing secret |
| `ADMIN_USERNAME` | `admin` | Initial admin username |
| `ADMIN_PASSWORD` | `changeme123` | Initial admin password |
| `UPLOAD_DIR` | `./uploads` | Upload directory |
| `MAX_UPLOAD_SIZE_MB` | `10` | Max upload file size (MB) |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | CORS allowed origins (comma-separated) |
