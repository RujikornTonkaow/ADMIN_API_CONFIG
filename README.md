# Portfolio Admin API

Go REST API สำหรับระบบจัดการหลายเว็บไซต์ โดยใช้ MongoDB เก็บข้อมูลแบบ multi-site และแยกข้อมูล portfolio ด้วย `site_id`

## Tech Stack

| Technology | Purpose |
|------------|---------|
| Go 1.22+ | API server ด้วย `net/http` ServeMux |
| MongoDB 7 | Database |
| JWT | Authentication |
| Docker | Local/dev deployment |

## Architecture

- Public Portfolio resolve site จาก domain ด้วย `GET /api/v1/public/sites/by-domain?host=:host`
- Public Portfolio ดึงข้อมูลด้วย `GET /api/v1/public/sites/{siteId}/portfolio`
- Admin Dashboard ใช้ site-scoped routes เช่น `/api/v1/admin/sites/{siteId}/portfolio/projects`
- `sites.domains` ใช้สำหรับ map domain/subdomain ไปยัง site
- `site_members` เป็น access list (`user_id` + `site_id`)
- Global role ใน JWT เป็นตัวตัดสินสิทธิ์: `super_admin`, `admin`, `editor`, `viewer`

## Quick Start

```bash
cp .env.example .env
docker compose up --build -d
```

หรือรัน local:

```bash
go mod tidy
go run ./cmd/server
```

API จะพร้อมใช้งานที่ `http://localhost:8080`

## Important Environment Variables

| Variable | Description |
|----------|-------------|
| `PORT` | API port |
| `MONGO_URI` | MongoDB connection string |
| `MONGO_DB` | Database name |
| `JWT_SECRET` | Secret สำหรับ sign JWT |
| `ADMIN_USERNAME` | Username ของ user seed เริ่มต้น |
| `ADMIN_PASSWORD` | Password ของ user seed เริ่มต้น |
| `ALLOWED_ORIGINS` | Static origins เช่น Admin Dashboard; portfolio domains โหลดจาก `sites.domains` |
| `RESET_DATABASE_ON_START` | ถ้า `true` จะ drop database แล้ว seed ใหม่ตอน start |

## Seed Data

เมื่อ database ว่าง ระบบจะ seed:

- user เริ่มต้นเป็น `super_admin`
- default portfolio site พร้อม domain `localhost:3000`
- access record ใน `site_members`
- default portfolio content ที่มี `site_id` ครบ

## Public API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/public/sites/by-domain?host=:host` | resolve site จาก host |
| GET | `/api/v1/public/sites/{siteId}/portfolio` | ดึง portfolio data ของ site |
| POST | `/api/v1/public/sites/{siteId}/portfolio/contacts` | ส่ง contact message |
| GET | `/uploads/{filename}` | ดูไฟล์ upload |

## Admin API

| Area | Path Pattern |
|------|--------------|
| Auth | `/api/v1/admin/auth/...` |
| Users | `/api/v1/admin/users...` |
| User site access | `/api/v1/admin/users/{id}/memberships` |
| Sites | `/api/v1/admin/sites...` |
| Portfolio content | `/api/v1/admin/sites/{siteId}/portfolio/...` |

Legacy portfolio admin routes เดิมไม่รองรับแล้ว ต้องใช้ site-scoped routes เท่านั้น

## Roles

| Role | สิทธิ์ |
|------|--------|
| `super_admin` | เห็นทุก site, จัดการ sites/users ทุกคน, สร้าง `super_admin` ได้ |
| `admin` | จัดการ users/content เฉพาะ site ที่ได้รับ access |
| `editor` | แก้ content เฉพาะ site ที่ได้รับ access |
| `viewer` | อ่าน Dashboard/Contacts เฉพาะ site ที่ได้รับ access |

## Docs

รายละเอียดเต็มอยู่ใน `GUIDE/`:

- `01-SYSTEM-FLOW.md`
- `04-API-REFERENCE.md`
- `05-DATABASE-SCHEMA.md`
- `06-SETUP-AND-DEPLOYMENT.md`
