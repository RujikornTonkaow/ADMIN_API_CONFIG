# System Flow — Portfolio Admin API

เอกสารนี้อธิบาย Flow การทำงานของระบบ **Portfolio Admin API** ตั้งแต่ Startup จนถึง Request/Response อย่างละเอียด

---

## สถาปัตยกรรมภาพรวม (Architecture Overview)

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Client Applications                         │
│  ┌──────────────────────┐       ┌──────────────────────────────┐   │
│  │  Portfolio Website   │       │     Admin Dashboard (SPA)    │   │
│  │  (Public visitor)    │       │     (Authenticated users)    │   │
│  └──────────┬───────────┘       └──────────────┬───────────────┘   │
└─────────────┼──────────────────────────────────┼───────────────────┘
              │ GET /api/v1/portfolio             │ JWT Bearer Token
              │ POST /api/v1/contact              │ CRUD operations
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
│  │  Handlers   │  │  Auth Middleware      │  │  RBAC Middleware  │  │
│  │  (Business  │  │  (JWT Validation)     │  │  (Role Check)    │  │
│  │   Logic)    │  │                       │  │                  │  │
│  └──────┬──────┘  └───────────────────────┘  └──────────────────┘  │
│         │                                                           │
│  ┌──────▼──────────────────────────────────────────────────────┐   │
│  │                   Repository Layer                          │   │
│  │   (Data Access — MongoDB CRUD operations)                   │   │
│  └──────────────────────────┬──────────────────────────────────┘   │
└─────────────────────────────┼──────────────────────────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │     MongoDB 7     │
                    │  (portfolio_admin) │
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
│  5. Ensure database indexes                                      │
│     └─→ สร้าง unique index บน admin_users.username               │
├──────────────────────────────────────────────────────────────────┤
│  6. Seed initial data (ถ้า admin_users collection ว่าง)           │
│     └─→ สร้าง admin user ด้วย bcrypt hashed password             │
│     └─→ สร้าง default: site_settings, hero, about,               │
│         skills, projects, experiences, social_links               │
├──────────────────────────────────────────────────────────────────┤
│  7. Initialize all repositories                                  │
│     └─→ สร้าง repository instance สำหรับแต่ละ collection          │
├──────────────────────────────────────────────────────────────────┤
│  8. Build router with middleware stack                            │
│     └─→ ลงทะเบียน routes ทั้งหมดพร้อม middleware                  │
├──────────────────────────────────────────────────────────────────┤
│  9. Start HTTP server                                            │
│     └─→ ReadTimeout: 15s, ReadHeaderTimeout: 5s                 │
│     └─→ WriteTimeout: 15s, IdleTimeout: 60s                     │
│     └─→ MaxHeaderBytes: 1MB                                     │
├──────────────────────────────────────────────────────────────────┤
│  10. Wait for shutdown signal (SIGINT / SIGTERM)                 │
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
  │  GET /api/v1/admin/skills      │
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
│  Level 3       │  admin         │  ทุกอย่าง + จัดการ    │
│  (สูงสุด)      │                │  User CRUD           │
├────────────────┼────────────────┼──────────────────────┤
│  Level 2       │  user_account  │  CRUD content ทั้งหมด │
│                │                │  Upload, ลบ Contact   │
├────────────────┼────────────────┼──────────────────────┤
│  Level 1       │  visitor       │  ดู content + อ่าน    │
│  (ต่ำสุด)      │                │  Contact messages     │
└────────────────┴────────────────┴──────────────────────┘

การตรวจสอบ: ถ้า user มี role level >= required level → ผ่าน
ตัวอย่าง: route ต้องการ user_account (2)
  - admin (3) → ผ่าน ✓
  - user_account (2) → ผ่าน ✓
  - visitor (1) → ไม่ผ่าน ✗ → 403
```

---

## 4. Public Portfolio Flow

เมื่อ Frontend เรียก `GET /api/v1/portfolio`:

```
┌────────────────────────────────────────────────────────────────┐
│  PublicHandler.GetPortfolio()                                  │
│                                                                │
│  1. Fetch site_settings   ──► MongoDB: site_settings collection│
│  2. Fetch hero            ──► MongoDB: hero collection         │
│  3. Fetch about           ──► MongoDB: about collection        │
│  4. Fetch skills[]        ──► MongoDB: skills (sorted)         │
│  5. Fetch projects[]      ──► MongoDB: projects (sorted)       │
│  6. Fetch experiences[]   ──► MongoDB: experiences (sorted)    │
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

```
                     ┌─────────────────────────────────────┐
                     │        Admin Dashboard (SPA)        │
                     └──────────────┬──────────────────────┘
                                    │
          ┌─────────────────────────┼─────────────────────────┐
          │                         │                         │
    GET /projects            POST /projects          PUT /projects/{id}
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
   │                  MongoDB: projects collection               │
   └─────────────────────────────────────────────────────────────┘

   เพิ่มเติม:
   - DELETE /projects/{id}     → ลบ project
   - PUT /projects/reorder     → เรียงลำดับใหม่ด้วย { ids: [...] }
```

---

## 6. File Upload Flow

```
Client                          API Server                    Disk
  │                                │                            │
  │  POST /api/v1/admin/upload     │                            │
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
  │  POST /api/v1/contact          │                            │
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
  │                                │  (is_read: false)          │
  │                                │ ──────────────────────────►│
  │                                │                            │
  │  201 Created                   │                            │
  │  { message: "sent" }           │                            │
  │◄────────────────────────────── │                            │

  Admin สามารถ:
  - GET  /admin/contacts       → ดูรายการทั้งหมด (visitor+)
  - GET  /admin/contacts/{id}  → ดูรายละเอียด (visitor+)
  - DELETE /admin/contacts/{id}→ ลบข้อความ (user_account+)
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
