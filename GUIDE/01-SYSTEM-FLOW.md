# System Flow — Portfolio Admin API

เอกสารนี้อธิบาย Flow การทำงานของระบบ **Portfolio Admin API** ตั้งแต่ Startup จนถึง Request/Response อย่างละเอียด

---

## Implementation Decision — ใช้ routes ใหม่เท่านั้น

ระบบนี้เลือก cutover เป็น multi-site เต็มรูปแบบ:

- ใช้ Portfolio admin routes แบบ `/api/v1/admin/sites/{siteId}/portfolio/...` เท่านั้น
- ไม่เก็บ legacy Portfolio admin routes เดิม
- ข้อมูล portfolio เดิมไม่ต้อง migrate ให้ตั้ง `RESET_DATABASE_ON_START=true` เพื่อ drop database แล้ว seed ใหม่ตอนรัน server
- ทุก handler/repository ของ portfolio ต้องรับ `siteId` จาก path/middleware และ query ด้วย `site_id` เสมอ

เอกสารที่ต้องอ่านคู่กันจากฝั่ง Admin Dashboard:

- `Website_Config/GUIDE/07-MULTI-SITE-ARCHITECTURE.md`
- `Website_Config/GUIDE/08-CORS-AND-LOCAL-DEV.md`

---

## สถาปัตยกรรมภาพรวม (Architecture Overview)

ระบบเป็นแบบ **multi-site**: เก็บ metadata ของแต่ละเว็บใน `sites` สมาชิกและบทบาทต่อไซต์ใน `site_members` และข้อมูล portfolio (hero, projects, …) ผูกกับ `site_id` เพื่อแยกขอบเขตต่อไซต์

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Client Applications                         │
│  ┌──────────────────────┐       ┌──────────────────────────────┐   │
│  │  Portfolio Website   │       │     Admin Dashboard (SPA)    │   │
│  │  (Public visitor)    │       │     (Authenticated users)    │   │
│  └──────────┬───────────┘       └──────────────┬───────────────┘   │
└─────────────┼──────────────────────────────────┼───────────────────┘
              │ GET /api/v1/public/sites/{siteId}/portfolio          │ JWT Bearer Token
              │ GET /api/v1/public/sites/by-domain?host=           │ CRUD ต่อ siteId
              │ POST .../portfolio/contacts                        │
              ▼                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Portfolio Admin API (Go)                        │
│                                                                     │
│  ┌─────────────────── Middleware Stack ───────────────────────┐     │
│  │  Recovery → CORS → Logging → RequestID → ServeMux         │     │
│  └───────────────────────────┬───────────────────────────────┘     │
│                              │                                      │
│  ┌───────────────────────────▼───────────────────────────────┐     │
│  │                     Router (ServeMux)                      │     │
│  │  Public Routes ──────── Auth Routes ──────── Admin Routes │     │
│  └──────┬──────────────────────┬─────────────────────┬───────┘     │
│         │                      │                     │              │
│  ┌──────▼──────┐  ┌───────────▼──────────┐  ┌───────▼──────────┐  │
│  │  Handlers   │  │  Auth Middleware      │  │  RBAC + Require  │  │
│  │  (Business  │  │  (JWT Validation)     │  │  SiteMember      │  │
│  │   Logic)    │  │                       │  │  (site-scoped)   │  │
│  └──────┬──────┘  └───────────────────────┘  └──────────────────┘  │
│         │                                                           │
│  ┌──────▼──────────────────────────────────────────────────────┐   │
│  │                   Repository Layer                          │   │
│  │   (Data Access — MongoDB — sites / site_members + ข้อมูลที่มี site_id) │
│  └──────────────────────────┬──────────────────────────────────┘   │
└─────────────────────────────┼──────────────────────────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │     MongoDB 7     │
                    │  (portfolio_admin) │
                    │  sites, site_members, │
                    │  collections อื่น ๆ (site-scoped) │
                    └───────────────────┘
```

---

## 1. Application Startup Flow

เมื่อ Server เริ่มทำงาน (`cmd/server/main.go`) จะทำตามลำดับนี้:

```
┌──────────────────────────────────────────────────────────────────┐
│  1. Initialize structured JSON logger (slog)                     │
│     └─→ Output format: JSON to stdout                            │
├──────────────────────────────────────────────────────────────────┤
│  2. Load configuration from environment variables                │
│     └─→ internal/config/config.go                                │
│     └─→ ถ้าไม่มี env var จะใช้ค่า default                         │
├──────────────────────────────────────────────────────────────────┤
│  3. Create upload directory                                      │
│     └─→ os.MkdirAll(UPLOAD_DIR) — สร้าง folder สำหรับเก็บรูป     │
├──────────────────────────────────────────────────────────────────┤
│  4. Connect to MongoDB                                           │
│     └─→ Timeout 10 วินาที                                        │
│     └─→ Ping เพื่อยืนยันการเชื่อมต่อ                              │
├──────────────────────────────────────────────────────────────────┤
│  5. Optional database reset                                      │
│     └─→ ถ้า RESET_DATABASE_ON_START=true จะ drop database ก่อน seed │
├──────────────────────────────────────────────────────────────────┤
│  6. Ensure database indexes                                      │
│     └─→ admin_users: unique index บน username                    │
│     └─→ sites: unique index บน slug                              │
│     └─→ site_members: unique compound index บน site_id + user_id │
├──────────────────────────────────────────────────────────────────┤
│  7. Seed initial data (ถ้า admin_users collection ว่าง)           │
│     └─→ สร้าง user เริ่มต้นเป็น super_admin ด้วย bcrypt hash     │
│     └─→ สร้าง default site                                      │
│     └─→ สร้าง site_members access record ให้ user เริ่มต้น        │
│     └─→ สร้างข้อมูล portfolio (site_settings, hero, about,       │
│         skills, projects, experiences, social_links) พร้อม site_id │
├──────────────────────────────────────────────────────────────────┤
│  8. Initialize all repositories                                  │
│     └─→ สร้าง repository instance สำหรับแต่ละ collection          │
├──────────────────────────────────────────────────────────────────┤
│  9. Build router with middleware stack                            │
│     └─→ ลงทะเบียน routes ทั้งหมดพร้อม middleware                  │
├──────────────────────────────────────────────────────────────────┤
│  10. Start HTTP server                                           │
│     └─→ ReadTimeout: 15s, ReadHeaderTimeout: 5s                 │
│     └─→ WriteTimeout: 15s, IdleTimeout: 60s                     │
│     └─→ MaxHeaderBytes: 1MB                                     │
├──────────────────────────────────────────────────────────────────┤
│  11. Wait for shutdown signal (SIGINT / SIGTERM)                 │
│      └─→ Graceful shutdown with 10s timeout                      │
│      └─→ ปิด DB connection อย่างสมบูรณ์ด้วย defer disconnect()    │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. Request Processing Flow

ทุก HTTP request ที่เข้ามาจะผ่าน Middleware Stack ตามลำดับ:

```
HTTP Request
    │
    ▼
┌─────────────────────────────┐
│  1. Recovery Middleware      │  ← ดักจับ panic ป้องกัน server crash
│     └─→ recover() + log     │     ส่ง 500 Internal Server Error
└─────────────┬───────────────┘
              ▼
┌─────────────────────────────┐
│  2. CORS Middleware          │  ← ตรวจสอบ Origin header
│     └─→ ถ้า OPTIONS → 204   │     set Access-Control-* headers
│     └─→ ตรวจ allowed origins │
└─────────────┬───────────────┘
              ▼
┌─────────────────────────────┐
│  3. Logging Middleware       │  ← บันทึก method, path, status, duration
│     └─→ structured JSON log │     แนบ request_id
└─────────────┬───────────────┘
              ▼
┌─────────────────────────────┐
│  4. RequestID Middleware     │  ← สร้าง UUID v4
│     └─→ ใส่ context + header │     X-Request-ID header
└─────────────┬───────────────┘
              ▼
┌─────────────────────────────┐
│  5. ServeMux (Router)        │  ← จับคู่ route กับ handler
│     └─→ Public / Auth /     │
│         Admin routes         │
└─────────────┬───────────────┘
              ▼
        Route Handler
```

---

## 3. Authentication & Authorization Flow

### 3.1 Login Flow

```
Client                          API Server                       MongoDB
  │                                │                                │
  │  POST /api/v1/admin/auth/login │                                │
  │  { username, password }        │                                │
  │ ──────────────────────────────►│                                │
  │                                │  FindByUsername(username)       │
  │                                │ ──────────────────────────────►│
  │                                │◄────────────────────────────── │
  │                                │  AdminUser document            │
  │                                │                                │
  │                                │  bcrypt.Compare(hash, password)│
  │                                │  ✓ Match                      │
  │                                │                                │
  │                                │  JWT Sign (HS256)              │
  │                                │  Claims: sub, usr, role, exp   │
  │                                │  Expiry: 24 hours              │
  │                                │                                │
  │  { token, user }               │                                │
  │◄────────────────────────────── │                                │
```

### 3.2 Protected Route Flow

```
Client                          API Server
  │                                │
  │  GET /api/v1/admin/sites/{siteId}/portfolio/skills │
  │  Authorization: Bearer <JWT>   │
  │ ──────────────────────────────►│
  │                                │
  │                         ┌──────┴──────────────────────────┐
  │                         │  Auth Middleware                 │
  │                         │  1. Extract "Bearer " token      │
  │                         │  2. jwt.Parse with HMAC secret   │
  │                         │  3. Extract claims: sub,usr,role │
  │                         │  4. Store in request context     │
  │                         └──────┬──────────────────────────┘
  │                                │
  │                         ┌──────┴──────────────────────────┐
  │                         │  RequireRole Middleware          │
  │                         │  1. Get role from context        │
  │                         │  2. Compare RoleLevel(user) >=   │
  │                         │     RoleLevel(required)          │
  │                         │  3. ✓ Pass → next handler        │
  │                         │     ✗ Fail → 403 Forbidden       │
  │                         └──────┬──────────────────────────┘
  │                                │
  │                           Handler Logic
  │                                │
  │  { data: [...] }               │
  │◄────────────────────────────── │
```

### 3.3 RBAC Role Hierarchy

```
┌────────────────────────────────────────────────────────┐
│  Role Level    │  Role Name     │  สิทธิ์ที่ทำได้      │
├────────────────┼────────────────┼──────────────────────┤
│  Level 4       │  super_admin   │  ทุกอย่าง ทุก site    │
│  (สูงสุด)      │                │  และสร้าง super admin │
├────────────────┼────────────────┼──────────────────────┤
│  Level 3       │  admin         │  User/content เฉพาะ   │
│                │                │  site ที่ได้รับสิทธิ์ │
├────────────────┼────────────────┼──────────────────────┤
│  Level 2       │  editor        │  แก้ content เฉพาะ    │
│                │                │  site ที่ได้รับสิทธิ์ │
├────────────────┼────────────────┼──────────────────────┤
│  Level 1       │  viewer        │  ดู Messages/Contacts │
│  (ต่ำสุด)      │                │  Contact messages     │
└────────────────┴────────────────┴──────────────────────┘

การตรวจสอบ: ถ้า user มี role level >= required level → ผ่าน
ตัวอย่าง: route ต้องการ editor (2)
  - admin (3) → ผ่าน ✓
  - editor (2) → ผ่าน ✓
  - viewer (1) → ไม่ผ่าน ✗ → 403
```

### 3.4 Site Membership Middleware (RequireSiteMember)

Route ฝั่ง admin ที่ผูกกับ `siteId` จะผ่าน middleware **RequireSiteMember** เพื่อตรวจว่า user ปัจจุบันมี access record ใน `site_members` หรือไม่ และใช้ global role เป็นตัวตัดสินความสามารถ:

```
site_members = user_id + site_id access list
super_admin = bypass site access
admin/editor/viewer = ต้องมี membership ของ site นั้น
```

**super_admin** ข้ามการตรวจสอบสมาชิกไซต์ ส่วน `admin`, `editor`, `viewer` ต้องมีแถวใน `site_members` สำหรับไซต์นั้น

---

## 4. Public Portfolio Flow

**แยกไซต์ตาม `siteId`:** Frontend อาจใช้ `GET /api/v1/public/sites/by-domain?host=<hostname>` เพื่อ resolve ว่า host นั้นชี้ไปที่ `siteId` ใด จากนั้นเรียก `GET /api/v1/public/sites/{siteId}/portfolio`

เมื่อ Frontend เรียก `GET /api/v1/public/sites/{siteId}/portfolio`:

```
┌────────────────────────────────────────────────────────────────┐
│  PublicHandler.GetPortfolio(siteId)                             │
│                                                                │
│  1. Fetch site_settings   ──► MongoDB: filter ด้วย site_id     │
│  2. Fetch hero            ──► MongoDB: hero (site-scoped)      │
│  3. Fetch about           ──► MongoDB: about (site-scoped)      │
│  4. Fetch skills[]        ──► MongoDB: skills (sorted, site_id) │
│  5. Fetch projects[]      ──► MongoDB: projects (sorted)       │
│  6. Fetch experiences[]   ──► MongoDB: experiences (sorted)     │
│  7. Fetch social_links[]  ──► MongoDB: social_links (sorted)   │
│  8. Build nav_items[]     ──► Hardcoded navigation anchors     │
│                                                                │
│  ⚠ Partial Failure: ถ้า fetch site_settings/hero/about ล้มเหลว│
│  ระบบจะ log error แต่ยังคง return 200 OK พร้อม zero values     │
│  สำหรับ section ที่ล้มเหลว (ไม่ crash ทั้ง response)            │
│  สำหรับ skills/projects/experiences/social_links ถ้า fetch      │
│  ล้มเหลว จะ return empty array [] แทน                          │
│                                                                │
│  Response: PortfolioData {                                     │
│    site_settings, hero, about, skills[],                       │
│    projects[], experiences[], social_links[], nav_items[]       │
│  }                                                             │
└────────────────────────────────────────────────────────────────┘
```

---

## 5. Admin CRUD Flow (ตัวอย่าง: Projects)

เส้นทางฝั่ง admin ผูกกับไซต์ เช่น `GET /api/v1/admin/sites/{siteId}/portfolio/projects` (และ collection อื่น ๆ ในกลุ่ม portfolio ใช้รูปแบบเดียวกัน)

```
                     ┌─────────────────────────────────────┐
                     │        Admin Dashboard (SPA)        │
                     └──────────────┬──────────────────────┘
                                    │
          ┌─────────────────────────┼─────────────────────────┐
          │                         │                         │
 GET .../sites/{siteId}/portfolio/projects
 POST .../sites/{siteId}/portfolio/projects
 PUT  .../sites/{siteId}/portfolio/projects/{id}
          │                         │                         │
          ▼                         ▼                         ▼
   ┌──────────────┐   ┌──────────────────┐   ┌────────────────────┐
   │  List all     │   │  Create new      │   │  Update existing   │
   │  projects     │   │  project         │   │  project by ID     │
   │  sorted by    │   │  (validate +     │   │  (validate +       │
   │  sort_order   │   │   insert)        │   │   FindOneAndUpdate)│
   └──────┬───────┘   └────────┬─────────┘   └──────────┬─────────┘
          │                    │                         │
          ▼                    ▼                         ▼
   ┌─────────────────────────────────────────────────────────────┐
   │         MongoDB: projects collection (filter ด้วย site_id)  │
   └─────────────────────────────────────────────────────────────┘

   เพิ่มเติม:
   - DELETE .../portfolio/projects/{id}     → ลบ project
   - PUT .../portfolio/projects/reorder     → เรียงลำดับใหม่ด้วย { ids: [...] }
```

---

## 6. File Upload Flow

```
Client                          API Server                    Disk
  │                                │                            │
  │  POST /api/v1/admin/sites/{siteId}/portfolio/upload          │
  │  Content-Type: multipart/form  │                            │
  │  Body: file=<image>            │                            │
  │ ──────────────────────────────►│                            │
  │                                │                            │
  │                         ┌──────┴───────────────────────┐    │
  │                         │  1. Check file size           │    │
  │                         │     <= MAX_UPLOAD_SIZE_MB     │    │
  │                         │  2. Validate file extension    │    │
  │                         │     (.jpg, .jpeg, .png, .gif, │    │
  │                         │      .webp, .svg)             │    │
  │                         │  3. Generate UUID filename    │    │
  │                         │     uuid + extension          │    │
  │                         │  4. Save to UPLOAD_DIR        │    │
  │                         └──────┬───────────────────────┘    │
  │                                │  Write file                │
  │                                │ ──────────────────────────►│
  │                                │                            │
  │  { url: "/uploads/uuid.jpg" }  │                            │
  │◄────────────────────────────── │                            │

  ไฟล์ที่ upload แล้วสามารถเข้าถึงได้ผ่าน:
  GET /uploads/{filename} → Static file server (ไม่ต้อง auth)
```

---

## 7. Contact Form Flow

```
Website Visitor                 API Server                   MongoDB
  │                                │                            │
  │  POST /api/v1/public/sites/{siteId}/portfolio/contacts     │
  │  { name, email,                │                            │
  │    subject, message }          │                            │
  │ ──────────────────────────────►│                            │
  │                                │                            │
  │                         ┌──────┴───────────────────────┐    │
  │                         │  Validate:                    │    │
  │                         │  - ทุก field ต้องไม่ว่าง       │    │
  │                         │  - email format ถูกต้อง       │    │
  │                         └──────┬───────────────────────┘    │
  │                                │                            │
  │                                │  Insert to                 │
  │                                │  contact_messages           │
  │                                │  (is_read: false, site_id)  │
  │                                │ ──────────────────────────►│
  │                                │                            │
  │  201 Created                   │                            │
  │  { message: "sent" }           │                            │
  │◄────────────────────────────── │                            │

  Admin สามารถ (ต่อไซต์):
  - GET  /api/v1/admin/sites/{siteId}/portfolio/contacts       → ดูรายการทั้งหมด (viewer+)
  - GET  /api/v1/admin/sites/{siteId}/portfolio/contacts/{id}  → ดูรายละเอียด (viewer+)
  - DELETE /api/v1/admin/sites/{siteId}/portfolio/contacts/{id}  → ลบข้อความ (editor+)
```

---

## 8. Graceful Shutdown Flow

```
┌──────────────────────────────────────────────────────────────┐
│  SIGINT (Ctrl+C) หรือ SIGTERM received                       │
│                                                              │
│  1. Log "shutdown signal received"                           │
│  2. สร้าง shutdown context (timeout 10 วินาที)                │
│  3. srv.Shutdown(ctx)                                        │
│     └─→ หยุดรับ connection ใหม่                               │
│     └─→ รอ in-flight requests เสร็จ (หรือ timeout)            │
│  4. defer disconnect() — ปิด MongoDB connection              │
│  5. Log "server stopped gracefully"                          │
└──────────────────────────────────────────────────────────────┘
```

---

## 9. JSON Response Format

API ทุก endpoint ส่ง response ในรูปแบบ Envelope:

```json
// Success
{
  "data": { ... },
  "meta": { ... }       // optional
}

// Error
{
  "error": "error message here"
}
```

---

## 10. Data Flow Summary

```
                    ┌──────────────────────┐
                    │    Environment Vars   │
                    │    (.env file)        │
                    └──────────┬───────────┘
                               │ config.Load()
                               ▼
┌──────────┐    ┌──────────────────────────┐    ┌──────────────┐
│  Client  │───►│      HTTP Server         │───►│   MongoDB    │
│  (SPA/   │    │                          │    │              │
│  Website)│◄───│  Middleware → Router     │◄───│  Collections │
└──────────┘    │  → Handler → Repository  │    └──────────────┘
                └────────────┬─────────────┘
                             │
                    ┌────────▼────────┐
                    │   Upload Dir    │
                    │  (./uploads)    │
                    └─────────────────┘
```
