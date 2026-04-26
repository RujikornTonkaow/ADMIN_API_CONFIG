# API Reference — Portfolio Admin API

เอกสารนี้อธิบาย API Endpoints ทั้งหมดอย่างละเอียด รวมถึง Request/Response format, Authentication, และตัวอย่างการใช้งาน

**การเป็นสมาชิกไซต์ (multi-site):** การอนุญาตแบ่งเป็น **สองระดับ** — **(1) Global RBAC** จาก JWT (`super_admin`, `admin`, `editor`, `viewer`) และ **(2) Site access** จาก collection `site_members` ต่อแต่ละ `siteId` ผู้ใช้ `super_admin` ข้ามการตรวจสอบ site access และเห็นทุก site ส่วน role อื่นต้องมี record ใน `site_members`

---

## Implementation Decision — Multi-site Cutover

โปรเจกต์ backend ต้องใช้ **site-scoped routes แบบใหม่เท่านั้น** สำหรับ Portfolio CRUD:

```text
/api/v1/admin/sites/{siteId}/portfolio/...
```

ไม่ต้องเก็บ legacy routes เดิม เช่น:

```text
/api/v1/admin/projects
/api/v1/admin/skills
/api/v1/admin/hero
/api/v1/admin/upload
```

ถ้ามี legacy routes เหลือใน router/handler/test ให้ลบออกหรือแก้ให้ test ใช้ route ใหม่ทั้งหมด เพื่อบังคับให้ Admin Dashboard ส่ง `siteId` เสมอ

ข้อมูล portfolio เดิมใน MongoDB ให้ถือว่าไม่ต้อง migrate: ให้ตั้ง `RESET_DATABASE_ON_START=true` ตอนรัน server เพื่อ drop database แล้ว seed ใหม่ สร้าง `sites`, `site_members` และ portfolio documents ที่มี `site_id` ครบ

เอกสารที่ฝั่ง backend ควรอ่านคู่กับไฟล์นี้:

- `Website_Config/GUIDE/07-MULTI-SITE-ARCHITECTURE.md`
- `Website_Config/GUIDE/08-CORS-AND-LOCAL-DEV.md`

---

## Base URL

```
http://localhost:8080/api/v1
```

---

## Response Format (Envelope)

ทุก endpoint ส่ง response ในรูปแบบ JSON Envelope:

```json
// Success
{
  "data": { ... }
}

// Success with metadata (เช่น contacts list)
{
  "data": [ ... ],
  "meta": { "unread_count": 3 }
}

// Error
{
  "error": "error message"
}
```

**หมายเหตุสำคัญ:** API ใช้ `DisallowUnknownFields` — ถ้าส่ง JSON field ที่ไม่มีอยู่ใน struct จะได้ `400 Bad Request` ทันที

---

## Authentication

Endpoints ที่ต้องการ authentication ต้องส่ง JWT token ใน header:

```
Authorization: Bearer <jwt_token>
```

Token ได้จาก `POST /api/v1/admin/auth/login` มีอายุ 24 ชั่วโมง

---

## Role Requirements

| สัญลักษณ์ | Role ขั้นต่ำ | ใครเข้าได้บ้าง |
|-----------|-------------|---------------|
| Public | ไม่ต้อง auth | ทุกคน |
| viewer+ | viewer | viewer, editor, admin, super_admin |
| editor+ | editor | editor, admin, super_admin |
| admin+ | admin | admin, super_admin |
| super_admin | super_admin | super_admin เท่านั้น |

### Site access (ใช้คู่กับ JWT สำหรับ route ภายใต้ `/api/v1/admin/sites/{siteId}/...`)

`site_members` เป็น access list ว่า user เข้าถึง site ไหนได้บ้าง ไม่ได้ใช้เป็น role ต่อ site แล้ว

- `super_admin` bypass site access และเห็นทุก site
- `admin`, `editor`, `viewer` ต้องมี record ใน `site_members` ของ site นั้น
- ความสามารถอ่าน/เขียนตัดสินจาก global role: `editor+` แก้ content, `viewer` เห็นเฉพาะ messages

---

## 1. Public API (ไม่ต้อง Auth)

### GET /api/v1/public/sites/by-domain

แปลง hostname เป็น site (ใช้ตอน frontend รู้แค่ domain ยังไม่รู้ `siteId`)

**Auth:** ไม่ต้อง

**Query:** `host` (required) — hostname ที่ต้องการ resolve (เช่น `www.example.com`)

**Response:** `200 OK` — object ของ site ใน `data`

**Errors:**
- `400` — ไม่ส่ง `host` หรือค่าว่าง
- `404` — ไม่พบไซต์ที่ผูก domain นี้

---

### GET /api/v1/public/sites/{siteId}/portfolio

ดึงข้อมูล portfolio ทั้งหมดของไซต์ที่ระบุ สำหรับแสดงในหน้าเว็บ

**Auth:** ไม่ต้อง

**Response:** `200 OK`
```json
{
  "data": {
    "site_settings": {
      "id": "...",
      "site_title": "Portfolio",
      "page_title": "Portfolio | Full-Stack Developer",
      "meta_description": "...",
      "footer_tagline": "Crafting digital experiences",
      "default_theme": "midnight",
      "profile_image": "/images/profile.jpg",
      "updated_at": "2025-01-15T10:00:00Z"
    },
    "hero": {
      "id": "...",
      "greeting": "Hello, I'm",
      "full_name": "John Doe",
      "subtitle": "Full-Stack Developer...",
      "cta_primary_text": "View My Work",
      "cta_primary_link": "#projects",
      "cta_secondary_text": "Get in Touch",
      "cta_secondary_link": "#contact",
      "updated_at": "2025-01-15T10:00:00Z"
    },
    "about": {
      "id": "...",
      "title": "Passionate about building great software",
      "bio_paragraphs": ["paragraph 1", "paragraph 2"],
      "personality_tags": ["Problem Solver", "Team Player"],
      "stats": [
        { "value": "5+", "label": "Years Experience" }
      ],
      "updated_at": "2025-01-15T10:00:00Z"
    },
    "skills": [
      {
        "id": "...",
        "name": "Vue.js",
        "icon": "logos:vue",
        "category": "frontend",
        "sort_order": 0,
        "created_at": "...",
        "updated_at": "..."
      }
    ],
    "projects": [
      {
        "id": "...",
        "title": "E-Commerce Platform",
        "description": "...",
        "tags": ["Nuxt 3", "Go", "MongoDB"],
        "image": "/uploads/xxx.jpg",
        "live_url": "https://...",
        "source_url": "https://github.com/...",
        "sort_order": 0,
        "created_at": "...",
        "updated_at": "..."
      }
    ],
    "experiences": [
      {
        "id": "...",
        "role": "Senior Developer",
        "company": "Tech Corp",
        "period": "2024 - Present",
        "description": "...",
        "highlights": ["highlight 1", "highlight 2"],
        "sort_order": 0,
        "created_at": "...",
        "updated_at": "..."
      }
    ],
    "social_links": [
      {
        "id": "...",
        "name": "GitHub",
        "url": "https://github.com/...",
        "icon": "mdi:github",
        "sort_order": 0,
        "created_at": "...",
        "updated_at": "..."
      }
    ],
    "nav_items": [
      { "label": "Home", "href": "#hero" },
      { "label": "About", "href": "#about" },
      { "label": "Skills", "href": "#skills" },
      { "label": "Projects", "href": "#projects" },
      { "label": "Experience", "href": "#experience" },
      { "label": "Contact", "href": "#contact" }
    ]
  }
}
```

---

### POST /api/v1/public/sites/{siteId}/portfolio/contacts

ส่งข้อความ contact จากผู้เยี่ยมชม (ผูกกับไซต์ที่ระบุ)

**Auth:** ไม่ต้อง

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "subject": "Project Inquiry",
  "message": "I'd like to discuss a project..."
}
```

**Validation:**
- ทุก field ต้องไม่ว่าง
- `email` ต้องมี format ถูกต้อง (มี `@` และ `.`)

**Response:** `201 Created`
```json
{
  "data": {
    "message": "Contact message sent successfully"
  }
}
```

**Errors:**
- `400` — field ว่างหรือ email format ผิด

---

## 2. Auth API

### POST /api/v1/admin/auth/login

เข้าสู่ระบบเพื่อรับ JWT token

**Auth:** ไม่ต้อง

**Request Body:**
```json
{
  "username": "admin",
  "password": "changeme123"
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "6789abcdef012345",
      "username": "admin",
      "role": "admin"
    }
  }
}
```

**Errors:**
- `400` — username หรือ password ว่าง
- `401` — credentials ไม่ถูกต้อง

---

### GET /api/v1/admin/auth/me

ดูข้อมูล user ปัจจุบัน (จาก JWT token)

**Auth:** viewer+

**Response:** `200 OK`
```json
{
  "data": {
    "id": "6789abcdef012345",
    "username": "admin",
    "role": "admin",
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-01-15T10:00:00Z"
  }
}
```

---

## 3. User Management API (admin+)

- `super_admin` จัดการผู้ใช้ได้ทั้งหมด และสร้าง `super_admin` คนอื่นได้
- `admin` จัดการได้เฉพาะ user ที่ share site access กัน และสร้างได้เฉพาะ `admin`, `editor`, `viewer`
- `editor` และ `viewer` ไม่มีสิทธิ์เข้า User Management

### GET /api/v1/admin/users

ดูรายการ users ทั้งหมด

**Auth:** admin

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "...",
      "username": "admin",
      "role": "admin",
      "created_at": "...",
      "updated_at": "..."
    },
    {
      "id": "...",
      "username": "editor",
      "role": "editor",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

---

### POST /api/v1/admin/users

สร้าง user ใหม่

**Auth:** admin

**Request Body:**
```json
{
  "username": "editor",
  "password": "securepassword",
  "role": "editor"
}
```

**Validation:**
- `username` — ต้องไม่ว่าง, ต้องไม่ซ้ำ (unique index)
- `password` — ต้องไม่ว่าง, ขั้นต่ำ 8 ตัวอักษร
- `role` — ต้องเป็น `super_admin`, `admin`, `editor`, หรือ `viewer`

**Response:** `201 Created` — return user object (password ถูกซ่อนจาก JSON)

**Errors:**
- `400` — validation ไม่ผ่าน
- `409` — username ซ้ำ

---

### GET /api/v1/admin/users/{id}

ดูข้อมูล user ตาม ID

**Auth:** admin

**Response:** `200 OK`

**Errors:**
- `400` — ID format ไม่ถูกต้อง
- `404` — ไม่พบ user

---

### PUT /api/v1/admin/users/{id}

แก้ไขข้อมูล user (username, role)

**Auth:** admin

**Request Body:**
```json
{
  "username": "new_username",
  "role": "editor"
}
```

**Validation:**
- `username` และ `role` ต้องไม่ว่าง
- `role` ต้องเป็น `super_admin`, `admin`, `editor`, หรือ `viewer`

**Safety Rules:**
- ไม่สามารถ downgrade role ของตัวเอง (ถ้าเป็น admin แล้วเปลี่ยนเป็น role อื่น)
- ไม่สามารถ downgrade admin คนสุดท้ายออกจาก role admin (ต้องมี admin อย่างน้อย 1 คนเสมอ)

**Response:** `200 OK` — return updated user object

**Errors:**
- `400` — validation ไม่ผ่าน, downgrade ตัวเอง, หรือลบ admin คนสุดท้าย
- `404` — ไม่พบ user
- `409` — username ซ้ำ

---

### DELETE /api/v1/admin/users/{id}

ลบ user

**Auth:** admin

**Safety Rules:**
- ไม่สามารถลบ account ตัวเอง
- ไม่สามารถลบ admin คนสุดท้าย (ต้องมี admin อย่างน้อย 1 คนเสมอ)

**Response:** `204 No Content` (ไม่มี body)

**Errors:**
- `400` — ลบตัวเอง, หรือลบ admin คนสุดท้าย
- `404` — ไม่พบ user

---

### PUT /api/v1/admin/users/{id}/password

เปลี่ยนรหัสผ่าน user

**Auth:** admin

**Request Body:**
```json
{
  "new_password": "newsecurepassword"
}
```

**Validation:**
- `new_password` ขั้นต่ำ 8 ตัวอักษร

**Response:** `200 OK`
```json
{
  "data": {
    "status": "password changed"
  }
}
```

**Errors:**
- `400` — password สั้นเกินไป
- `404` — ไม่พบ user

---

## 4. Site Management API

```
GET    /api/v1/admin/sites              — ลิสต์ไซต์ที่ user เห็นได้ (viewer+)
POST   /api/v1/admin/sites              — สร้างไซต์ใหม่ (super_admin เท่านั้น)
GET    /api/v1/admin/sites/{siteId}     — ดูรายละเอียดไซต์ (site viewer+)
PUT    /api/v1/admin/sites/{siteId}     — แก้ไขไซต์ (site owner)
DELETE /api/v1/admin/sites/{siteId}     — ลบไซต์ (site owner)
```

### GET /api/v1/admin/sites

ลิสต์ไซต์ทั้งหมดที่ผู้ใช้ปัจจุบันเป็นสมาชิก

**Auth:** viewer+

**Response:** `200 OK` — array ของ site ใน `data`

---

### POST /api/v1/admin/sites

สร้างไซต์ใหม่ ผู้สร้างจะถูกเพิ่มเป็น `owner` ใน `site_members` อัตโนมัติ

**Auth:** super_admin

**Request Body:**
```json
{
  "name": "My Portfolio",
  "slug": "my-portfolio",
  "type": "portfolio",
  "domains": ["example.com", "www.example.com"]
}
```

- `type` — `portfolio` | `shop` | `finance`
- `domains` — array ของ hostname (ว่างได้ตาม validation ของ API)

**Response:** `201 Created` — object ของ site ใน `data`

**Errors:**
- `400` — validation ไม่ผ่าน

---

### GET /api/v1/admin/sites/{siteId}

ดูรายละเอียดไซต์

**Auth:** site viewer+ (สมาชิกไซต์ role ≥ `viewer`; หรือ global `admin`)

**Errors:**
- `400` — `siteId` ไม่ใช่ ObjectID ที่ถูกต้อง
- `403` — ไม่ใช่สมาชิกไซต์ (ยกเว้น `admin`)
- `404` — ไม่พบไซต์

---

### PUT /api/v1/admin/sites/{siteId}

แก้ไขข้อมูลไซต์

**Auth:** site owner (หรือ global `admin`)

**Request Body:**
```json
{
  "name": "Updated Name",
  "slug": "updated-slug",
  "domains": ["new.example.com"]
}
```

**Response:** `200 OK`

**Errors:**
- `400` / `403` / `404` — ตามกรณี

---

### DELETE /api/v1/admin/sites/{siteId}

ลบไซต์

**Auth:** site owner (หรือ global `admin`)

**Response:** `204 No Content`

**Errors:**
- `400` / `403` / `404` — ตามกรณี

---

## 5. Site Members API

```
GET    /api/v1/admin/sites/{siteId}/members                    — ลิสต์สมาชิก (site viewer+)
POST   /api/v1/admin/sites/{siteId}/members                    — เพิ่มสมาชิก (site owner)
PUT    /api/v1/admin/sites/{siteId}/members/{memberId}       — เปลี่ยน role (site owner)
DELETE /api/v1/admin/sites/{siteId}/members/{memberId}        — ลบสมาชิก (site owner)
```

### GET /api/v1/admin/sites/{siteId}/members

ลิสต์สมาชิกทั้งหมดของไซต์

**Auth:** site viewer+ (หรือ global `admin`)

**Response:** `200 OK`

---

### POST /api/v1/admin/sites/{siteId}/members

เพิ่มสมาชิกให้ไซต์

**Auth:** site owner (หรือ global `admin`)

**Request Body:**
```json
{
  "user_id": "6789abcdef0123456789abcd",
  "role": "editor"
}
```

- `role` — `owner` | `editor` | `viewer`

**Response:** `201 Created`

**Errors:**
- `400` — validation ไม่ผ่าน
- `403` / `404` — ตามกรณี

---

### PUT /api/v1/admin/sites/{siteId}/members/{memberId}

เปลี่ยน role ของสมาชิก

**Auth:** site owner (หรือ global `admin`)

**Request Body:**
```json
{
  "role": "viewer"
}
```

**Response:** `200 OK`

---

### DELETE /api/v1/admin/sites/{siteId}/members/{memberId}

ลบสมาชิกออกจากไซต์

**Auth:** site owner (หรือ global `admin`)

**Response:** `204 No Content`

---

## 6. Site Settings API

### GET /api/v1/admin/sites/{siteId}/portfolio/site-settings

ดู site settings ของไซต์ที่ระบุ

**Auth:** site viewer+ (หรือ global `admin`)

**Response:** `200 OK`
```json
{
  "data": {
    "id": "...",
    "site_title": "Portfolio",
    "page_title": "Portfolio | Full-Stack Developer",
    "meta_description": "...",
    "footer_tagline": "Crafting digital experiences",
    "default_theme": "midnight",
    "profile_image": "/images/profile.jpg",
    "updated_at": "..."
  }
}
```

---

### PUT /api/v1/admin/sites/{siteId}/portfolio/site-settings

แก้ไข site settings

**Auth:** site editor+ (หรือ global `admin`)

**Request Body:**
```json
{
  "site_title": "My Portfolio",
  "page_title": "My Portfolio | Developer",
  "meta_description": "Developer portfolio",
  "footer_tagline": "Building the web",
  "default_theme": "midnight",
  "profile_image": "/uploads/profile.jpg"
}
```

**Validation:**
- `site_title` ต้องไม่ว่าง

**Response:** `200 OK`

---

## 7. Hero Section API

### GET /api/v1/admin/sites/{siteId}/portfolio/hero

ดูข้อมูล hero section

**Auth:** site viewer+ (หรือ global `admin`)

---

### PUT /api/v1/admin/sites/{siteId}/portfolio/hero

แก้ไข hero section

**Auth:** site editor+ (หรือ global `admin`)

**Request Body:**
```json
{
  "greeting": "Hello, I'm",
  "full_name": "John Doe",
  "subtitle": "Full-Stack Developer",
  "cta_primary_text": "View My Work",
  "cta_primary_link": "#projects",
  "cta_secondary_text": "Get in Touch",
  "cta_secondary_link": "#contact"
}
```

**Validation:**
- `full_name` ต้องไม่ว่าง

---

## 8. About Section API

### GET /api/v1/admin/sites/{siteId}/portfolio/about

ดูข้อมูล about section

**Auth:** site viewer+ (หรือ global `admin`)

---

### PUT /api/v1/admin/sites/{siteId}/portfolio/about

แก้ไข about section

**Auth:** site editor+ (หรือ global `admin`)

**Request Body:**
```json
{
  "title": "About Me",
  "bio_paragraphs": [
    "First paragraph...",
    "Second paragraph..."
  ],
  "personality_tags": ["Problem Solver", "Team Player"],
  "stats": [
    { "value": "5+", "label": "Years Experience" },
    { "value": "30+", "label": "Projects Completed" }
  ]
}
```

**Validation:**
- `title` ต้องไม่ว่าง

---

## 9. Skills API

### GET /api/v1/admin/sites/{siteId}/portfolio/skills

ดูรายการ skills ทั้งหมด (เรียงตาม sort_order)

**Auth:** site viewer+ (หรือ global `admin`)

---

### POST /api/v1/admin/sites/{siteId}/portfolio/skills

สร้าง skill ใหม่

**Auth:** site editor+ (หรือ global `admin`)

**Request Body:**
```json
{
  "name": "React",
  "icon": "logos:react",
  "category": "frontend",
  "sort_order": 5
}
```

**Validation:**
- `name`, `icon`, `category` ต้องไม่ว่าง
- `category` ต้องเป็นหนึ่งใน: `frontend`, `backend`, `devops`, `tools`

**Response:** `201 Created`

**หมายเหตุ:** Skills ไม่มี reorder endpoint — `sort_order` กำหนดตอนสร้างเท่านั้น

---

### PUT /api/v1/admin/sites/{siteId}/portfolio/skills/{id}

แก้ไข skill

**Auth:** site editor+ (หรือ global `admin`)

**Validation:**
- `name`, `icon`, `category` ต้องไม่ว่าง

**Response:** `200 OK`

**Errors:**
- `404` — ไม่พบ skill

---

### DELETE /api/v1/admin/sites/{siteId}/portfolio/skills/{id}

ลบ skill

**Auth:** site editor+ (หรือ global `admin`)

**Response:** `204 No Content` (ไม่มี body)

**Errors:**
- `404` — ไม่พบ skill

---

## 10. Projects API

### GET /api/v1/admin/sites/{siteId}/portfolio/projects

ดูรายการ projects ทั้งหมด (เรียงตาม sort_order)

**Auth:** site viewer+ (หรือ global `admin`)

---

### POST /api/v1/admin/sites/{siteId}/portfolio/projects

สร้าง project ใหม่

**Auth:** site editor+ (หรือ global `admin`)

**Request Body:**
```json
{
  "title": "My App",
  "description": "A web application...",
  "tags": ["Vue 3", "Go", "MongoDB"],
  "image": "/uploads/project.jpg",
  "live_url": "https://myapp.com",
  "source_url": "https://github.com/user/repo",
  "sort_order": 0
}
```

**Validation:**
- `title` และ `description` ต้องไม่ว่าง
- `tags` ถ้าไม่ส่งจะ default เป็น `[]`

**Response:** `201 Created`

---

### PUT /api/v1/admin/sites/{siteId}/portfolio/projects/{id}

แก้ไข project

**Auth:** site editor+ (หรือ global `admin`)

**Validation:**
- `title` และ `description` ต้องไม่ว่าง

**Response:** `200 OK`

**Errors:**
- `404` — ไม่พบ project

---

### DELETE /api/v1/admin/sites/{siteId}/portfolio/projects/{id}

ลบ project

**Auth:** site editor+ (หรือ global `admin`)

**Response:** `204 No Content` (ไม่มี body)

**Errors:**
- `404` — ไม่พบ project

---

### PUT /api/v1/admin/sites/{siteId}/portfolio/projects/reorder

เรียงลำดับ projects ใหม่

**Auth:** site editor+ (หรือ global `admin`)

**Request Body:**
```json
{
  "ids": [
    "project_id_3",
    "project_id_1",
    "project_id_2"
  ]
}
```

**Validation:**
- `ids` ต้องไม่เป็น array ว่าง
- แต่ละ ID ต้องเป็น valid ObjectID format

ระบบจะกำหนด `sort_order` ตามตำแหน่งใน array (index 0 = sort_order 0, index 1 = sort_order 1, ...)

**Response:** `200 OK`
```json
{
  "data": {
    "status": "reordered"
  }
}
```

---

## 11. Experiences API

### GET /api/v1/admin/sites/{siteId}/portfolio/experiences

ดูรายการ experiences ทั้งหมด (เรียงตาม sort_order)

**Auth:** site viewer+ (หรือ global `admin`)

---

### POST /api/v1/admin/sites/{siteId}/portfolio/experiences

สร้าง experience ใหม่

**Auth:** site editor+ (หรือ global `admin`)

**Request Body:**
```json
{
  "role": "Senior Developer",
  "company": "Tech Corp",
  "period": "2024 - Present",
  "description": "Leading development...",
  "highlights": [
    "Built microservices",
    "Reduced deploy time by 60%"
  ],
  "sort_order": 0
}
```

**Validation:**
- `role`, `company`, `period` ต้องไม่ว่าง
- `highlights` ถ้าไม่ส่งจะ default เป็น `[]`

**Response:** `201 Created`

---

### PUT /api/v1/admin/sites/{siteId}/portfolio/experiences/{id}

แก้ไข experience

**Auth:** site editor+ (หรือ global `admin`)

**Validation:**
- `role`, `company`, `period` ต้องไม่ว่าง

**Response:** `200 OK`

**Errors:**
- `404` — ไม่พบ experience

---

### DELETE /api/v1/admin/sites/{siteId}/portfolio/experiences/{id}

ลบ experience

**Auth:** site editor+ (หรือ global `admin`)

**Response:** `204 No Content` (ไม่มี body)

**Errors:**
- `404` — ไม่พบ experience

---

### PUT /api/v1/admin/sites/{siteId}/portfolio/experiences/reorder

เรียงลำดับ experiences ใหม่

**Auth:** site editor+ (หรือ global `admin`)

**Request Body:**
```json
{
  "ids": ["exp_id_2", "exp_id_1", "exp_id_3"]
}
```

**Validation:**
- `ids` ต้องไม่เป็น array ว่าง
- แต่ละ ID ต้องเป็น valid ObjectID format

**Response:** `200 OK`
```json
{
  "data": {
    "status": "reordered"
  }
}
```

---

## 12. Social Links API

### GET /api/v1/admin/sites/{siteId}/portfolio/social-links

ดูรายการ social links ทั้งหมด (เรียงตาม sort_order)

**Auth:** site viewer+ (หรือ global `admin`)

---

### POST /api/v1/admin/sites/{siteId}/portfolio/social-links

สร้าง social link ใหม่

**Auth:** site editor+ (หรือ global `admin`)

**Request Body:**
```json
{
  "name": "GitHub",
  "url": "https://github.com/username",
  "icon": "mdi:github",
  "sort_order": 0
}
```

**Validation:**
- `name`, `url`, `icon` ต้องไม่ว่าง

**Response:** `201 Created`

---

### PUT /api/v1/admin/sites/{siteId}/portfolio/social-links/{id}

แก้ไข social link

**Auth:** site editor+ (หรือ global `admin`)

**Validation:**
- `name`, `url`, `icon` ต้องไม่ว่าง

**Response:** `200 OK`

**Errors:**
- `404` — ไม่พบ social link

---

### DELETE /api/v1/admin/sites/{siteId}/portfolio/social-links/{id}

ลบ social link

**Auth:** site editor+ (หรือ global `admin`)

**Response:** `204 No Content` (ไม่มี body)

**Errors:**
- `404` — ไม่พบ social link

---

### PUT /api/v1/admin/sites/{siteId}/portfolio/social-links/reorder

เรียงลำดับ social links ใหม่

**Auth:** site editor+ (หรือ global `admin`)

**Request Body:**
```json
{
  "ids": ["link_id_1", "link_id_3", "link_id_2"]
}
```

**Validation:**
- `ids` ต้องไม่เป็น array ว่าง
- แต่ละ ID ต้องเป็น valid ObjectID format

**Response:** `200 OK`
```json
{
  "data": {
    "status": "reordered"
  }
}
```

---

## 13. Contact Messages API

### GET /api/v1/admin/sites/{siteId}/portfolio/contacts

ดูรายการ contact messages ทั้งหมดของไซต์

**Auth:** site viewer+ (หรือ global `admin`)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "...",
      "name": "John Doe",
      "email": "john@example.com",
      "subject": "Project Inquiry",
      "message": "I'd like to...",
      "is_read": false,
      "created_at": "2025-01-15T10:00:00Z"
    }
  ],
  "meta": {
    "unread_count": 3
  }
}
```

---

### GET /api/v1/admin/sites/{siteId}/portfolio/contacts/{id}

ดูรายละเอียด contact message

**Auth:** site viewer+ (หรือ global `admin`)

**Side Effect:** ถ้าข้อความยังไม่ได้อ่าน (`is_read: false`) ระบบจะ auto mark เป็น `is_read: true`

**Response:** `200 OK`

**Errors:**
- `400` — ID format ไม่ถูกต้อง
- `404` — ไม่พบ message

---

### DELETE /api/v1/admin/sites/{siteId}/portfolio/contacts/{id}

ลบ contact message

**Auth:** site editor+ (หรือ global `admin`)

**Response:** `204 No Content` (ไม่มี body)

**Errors:**
- `400` — ID format ไม่ถูกต้อง
- `404` — ไม่พบ message

---

## 14. File Upload API

### POST /api/v1/admin/sites/{siteId}/portfolio/upload

อัปโหลดไฟล์รูปภาพ (ผูกกับไซต์ที่ระบุ)

**Auth:** site editor+ (หรือ global `admin`)

**Request:**
- Content-Type: `multipart/form-data`
- Field name: `file`
- Max size: `MAX_UPLOAD_SIZE_MB` (default: 10 MB)

**Allowed file extensions:**
- `.jpg`, `.jpeg`
- `.png`
- `.gif`
- `.webp`
- `.svg`

**หมายเหตุ:** Validation เป็นแบบ file extension (ไม่ใช่ MIME type)

**Response:** `201 Created`
```json
{
  "data": {
    "url": "/uploads/a1b2c3d4-e5f6-7890-abcd-ef1234567890.jpg",
    "filename": "a1b2c3d4-e5f6-7890-abcd-ef1234567890.jpg"
  }
}
```

**Errors:**
- `400` — ไม่มีไฟล์, ไฟล์ใหญ่เกินไป, หรือ file extension ไม่รองรับ

---

## 15. Static Files

### GET /uploads/{filename}

เข้าถึงไฟล์ที่อัปโหลดแล้ว

**Auth:** ไม่ต้อง (public)

**ตัวอย่าง:** `GET http://localhost:8080/uploads/a1b2c3d4.jpg`

---

## HTTP Status Codes ที่ใช้

| Status | ความหมาย | เมื่อไหร่ |
|--------|---------|----------|
| `200` | OK | ดึงข้อมูล, อัปเดต, หรือ reorder สำเร็จ |
| `201` | Created | สร้างข้อมูลใหม่สำเร็จ (POST), อัปโหลดไฟล์สำเร็จ |
| `204` | No Content | ลบข้อมูลสำเร็จ (DELETE), CORS preflight |
| `400` | Bad Request | Request body ไม่ถูกต้อง, validation ไม่ผ่าน, ID format ผิด, unknown JSON fields |
| `401` | Unauthorized | ไม่มี token, token ไม่ถูกต้อง/หมดอายุ, format ผิด |
| `403` | Forbidden | Global role หรือ site role ไม่เพียงพอ, หรือไม่ใช่สมาชิกไซต์ |
| `404` | Not Found | ไม่พบ resource ตาม ID |
| `409` | Conflict | ข้อมูลซ้ำ (เช่น username ซ้ำ) |
| `500` | Internal Server Error | Server error (logged พร้อม request_id) |

---

## Response Headers

| Header | ค่า | คำอธิบาย |
|--------|-----|---------|
| `Content-Type` | `application/json` | ทุก API response |
| `X-Request-ID` | UUID v4 | ใช้ trace request ในระบบ |
| `Access-Control-Allow-Origin` | ตาม ALLOWED_ORIGINS | CORS origin |
| `Access-Control-Allow-Methods` | `GET, POST, PUT, DELETE, OPTIONS` | CORS methods |
| `Access-Control-Allow-Headers` | `Content-Type, Authorization` | CORS headers |
| `Access-Control-Allow-Credentials` | `true` | CORS credentials |
| `Access-Control-Max-Age` | `86400` | CORS preflight cache (24h) |

---

## Quick Test ด้วย cURL

แทนที่ `<siteId>` ด้วย ObjectID ของไซต์ (hex 24 ตัว)

```bash
# 1. Login
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"changeme123"}'

# 2. Resolve site จาก domain (public)
curl "http://localhost:8080/api/v1/public/sites/by-domain?host=localhost"

# 3. ดึง portfolio (public)
curl "http://localhost:8080/api/v1/public/sites/<siteId>/portfolio"

# 4. ดู skills (ต้อง auth + สมาชิกไซต์ระดับ viewer+)
curl "http://localhost:8080/api/v1/admin/sites/<siteId>/portfolio/skills" \
  -H "Authorization: Bearer <token>"

# 5. สร้าง skill ใหม่ (editor+)
curl -X POST "http://localhost:8080/api/v1/admin/sites/<siteId>/portfolio/skills" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"React","icon":"logos:react","category":"frontend","sort_order":5}'

# 6. อัปโหลดรูป (editor+)
curl -X POST "http://localhost:8080/api/v1/admin/sites/<siteId>/portfolio/upload" \
  -H "Authorization: Bearer <token>" \
  -F "file=@./photo.jpg"

# 7. ส่ง contact message (public)
curl -X POST "http://localhost:8080/api/v1/public/sites/<siteId>/portfolio/contacts" \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane","email":"jane@example.com","subject":"Hello","message":"Hi there!"}'
```
