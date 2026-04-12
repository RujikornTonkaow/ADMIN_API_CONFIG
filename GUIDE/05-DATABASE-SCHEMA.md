# Database Schema — Portfolio Admin API

เอกสารนี้อธิบายโครงสร้าง Database ทั้งหมด รวมถึง Collections, Schema, Indexes, และความสัมพันธ์ของข้อมูล

---

## Database Overview

| หัวข้อ | รายละเอียด |
|-------|-----------|
| **Database** | MongoDB 7 |
| **Database Name** | `portfolio_admin` (configurable via `MONGO_DB` env) |
| **Driver** | `go.mongodb.org/mongo-driver` v1.17.9 |
| **ID Format** | MongoDB ObjectID (`primitive.ObjectID`) |

---

## Collections ทั้งหมด

```
portfolio_admin (database)
├── admin_users          ← ข้อมูล users (มี unique index)
├── site_settings        ← Singleton: ตั้งค่า site
├── hero                 ← Singleton: Hero section
├── about                ← Singleton: About section
├── skills               ← Multiple: ทักษะ
├── projects             ← Multiple: ผลงาน
├── experiences          ← Multiple: ประสบการณ์
├── social_links         ← Multiple: ลิงก์โซเชียล
└── contact_messages     ← Multiple: ข้อความ contact
```

---

## 1. admin_users

ข้อมูลผู้ใช้ที่สามารถ login เข้าระบบ admin

### Schema

| Field | Type | BSON Tag | JSON Tag | คำอธิบาย |
|-------|------|----------|----------|---------|
| `_id` | `ObjectID` | `_id,omitempty` | `id` | Primary key (auto-generated) |
| `username` | `string` | `username` | `username` | ชื่อผู้ใช้ (unique) |
| `password` | `string` | `password` | `-` (hidden) | Password hash (bcrypt) |
| `role` | `string` | `role` | `role` | `admin` / `user_account` / `visitor` |
| `created_at` | `time.Time` | `created_at` | `created_at` | วันที่สร้าง |
| `updated_at` | `time.Time` | `updated_at` | `updated_at` | วันที่แก้ไขล่าสุด |

### Indexes

| Index | Fields | Type | คำอธิบาย |
|-------|--------|------|---------|
| `_id_` | `_id` | Default | MongoDB default primary key |
| `username_1` | `username` | **Unique** | ป้องกัน username ซ้ำ |

### ตัวอย่าง Document

```json
{
  "_id": ObjectId("6789abcdef012345abcdef12"),
  "username": "admin",
  "password": "$2a$10$...(bcrypt hash)...",
  "role": "admin",
  "created_at": ISODate("2025-01-15T10:00:00Z"),
  "updated_at": ISODate("2025-01-15T10:00:00Z")
}
```

### Role Values

| Role | Level | คำอธิบาย |
|------|-------|---------|
| `visitor` | 1 | สิทธิ์อ่านอย่างเดียว (ดู content, อ่าน contacts) |
| `user_account` | 2 | CRUD content, upload files, ลบ contacts |
| `admin` | 3 | ทุกอย่าง + จัดการ users |

---

## 2. site_settings (Singleton)

ตั้งค่าทั่วไปของเว็บไซต์ — มี 1 document เท่านั้น

### Schema

| Field | Type | BSON Tag | คำอธิบาย |
|-------|------|----------|---------|
| `_id` | `ObjectID` | `_id,omitempty` | Primary key |
| `site_title` | `string` | `site_title` | ชื่อ site (แสดงใน header/navbar) |
| `page_title` | `string` | `page_title` | HTML `<title>` |
| `meta_description` | `string` | `meta_description` | SEO meta description |
| `footer_tagline` | `string` | `footer_tagline` | ข้อความใน footer |
| `default_theme` | `string` | `default_theme` | Theme ค่าเริ่มต้น |
| `profile_image` | `string` | `profile_image` | URL รูปโปรไฟล์ |
| `updated_at` | `time.Time` | `updated_at` | วันที่แก้ไขล่าสุด |

### ตัวอย่าง Document

```json
{
  "_id": ObjectId("..."),
  "site_title": "Portfolio",
  "page_title": "Portfolio | Full-Stack Developer",
  "meta_description": "Full-Stack Developer portfolio...",
  "footer_tagline": "Crafting digital experiences",
  "default_theme": "midnight",
  "profile_image": "/images/profile.jpg",
  "updated_at": ISODate("2025-01-15T10:00:00Z")
}
```

---

## 3. hero (Singleton)

ข้อมูล Hero section — ส่วนแรกของหน้าเว็บ มี 1 document เท่านั้น

### Schema

| Field | Type | BSON Tag | คำอธิบาย |
|-------|------|----------|---------|
| `_id` | `ObjectID` | `_id,omitempty` | Primary key |
| `greeting` | `string` | `greeting` | คำทักทาย เช่น "Hello, I'm" |
| `full_name` | `string` | `full_name` | ชื่อเต็ม |
| `subtitle` | `string` | `subtitle` | คำอธิบายสั้น ๆ |
| `cta_primary_text` | `string` | `cta_primary_text` | ข้อความปุ่มหลัก |
| `cta_primary_link` | `string` | `cta_primary_link` | ลิงก์ปุ่มหลัก |
| `cta_secondary_text` | `string` | `cta_secondary_text` | ข้อความปุ่มรอง |
| `cta_secondary_link` | `string` | `cta_secondary_link` | ลิงก์ปุ่มรอง |
| `updated_at` | `time.Time` | `updated_at` | วันที่แก้ไขล่าสุด |

### ตัวอย่าง Document

```json
{
  "_id": ObjectId("..."),
  "greeting": "Hello, I'm",
  "full_name": "Puvakorn Pannasirichard",
  "subtitle": "Full-Stack Developer crafting performant, scalable web applications",
  "cta_primary_text": "View My Work",
  "cta_primary_link": "#projects",
  "cta_secondary_text": "Get in Touch",
  "cta_secondary_link": "#contact",
  "updated_at": ISODate("2025-01-15T10:00:00Z")
}
```

---

## 4. about (Singleton)

ข้อมูล About section มี 1 document เท่านั้น

### Schema

| Field | Type | BSON Tag | คำอธิบาย |
|-------|------|----------|---------|
| `_id` | `ObjectID` | `_id,omitempty` | Primary key |
| `title` | `string` | `title` | หัวข้อ About |
| `bio_paragraphs` | `[]string` | `bio_paragraphs` | เนื้อหา bio แต่ละย่อหน้า |
| `personality_tags` | `[]string` | `personality_tags` | Tags บุคลิกภาพ |
| `stats` | `[]Stat` | `stats` | สถิติ (embedded documents) |
| `updated_at` | `time.Time` | `updated_at` | วันที่แก้ไขล่าสุด |

### Stat (Embedded Document)

| Field | Type | คำอธิบาย |
|-------|------|---------|
| `value` | `string` | ค่าสถิติ เช่น "5+", "30+", "99%" |
| `label` | `string` | ป้ายกำกับ เช่น "Years Experience" |

### ตัวอย่าง Document

```json
{
  "_id": ObjectId("..."),
  "title": "Passionate about building great software",
  "bio_paragraphs": [
    "I'm a Full-Stack Developer with a passion...",
    "When I'm not coding, you'll find me..."
  ],
  "personality_tags": ["Problem Solver", "Team Player", "Continuous Learner"],
  "stats": [
    { "value": "5+", "label": "Years Experience" },
    { "value": "30+", "label": "Projects Completed" },
    { "value": "15+", "label": "Happy Clients" },
    { "value": "99%", "label": "Client Satisfaction" }
  ],
  "updated_at": ISODate("2025-01-15T10:00:00Z")
}
```

---

## 5. skills

ทักษะต่าง ๆ — หลาย documents, เรียงตาม `sort_order`

### Schema

| Field | Type | BSON Tag | คำอธิบาย |
|-------|------|----------|---------|
| `_id` | `ObjectID` | `_id,omitempty` | Primary key |
| `name` | `string` | `name` | ชื่อ skill เช่น "Vue.js", "Go" |
| `icon` | `string` | `icon` | ชื่อ icon (Iconify format) เช่น "logos:vue" |
| `category` | `string` | `category` | หมวดหมู่: `frontend`, `backend`, `devops`, `tools` |
| `sort_order` | `int` | `sort_order` | ลำดับการแสดงผล |
| `created_at` | `time.Time` | `created_at` | วันที่สร้าง |
| `updated_at` | `time.Time` | `updated_at` | วันที่แก้ไขล่าสุด |

### Category Values

| Category | ตัวอย่าง Skills |
|----------|----------------|
| `frontend` | Vue.js, Nuxt, TypeScript, TailwindCSS |
| `backend` | Go, Node.js, MongoDB, Redis |
| `devops` | Docker, Kubernetes, GitHub Actions |
| `tools` | Git, Figma, VS Code |

---

## 6. projects

ผลงาน/โปรเจกต์ — หลาย documents, เรียงตาม `sort_order`

### Schema

| Field | Type | BSON Tag | คำอธิบาย |
|-------|------|----------|---------|
| `_id` | `ObjectID` | `_id,omitempty` | Primary key |
| `title` | `string` | `title` | ชื่อโปรเจกต์ |
| `description` | `string` | `description` | คำอธิบายโปรเจกต์ |
| `tags` | `[]string` | `tags` | แท็กเทคโนโลยี เช่น ["Vue 3", "Go"] |
| `image` | `string` | `image,omitempty` | URL รูปภาพ (optional) |
| `live_url` | `string` | `live_url,omitempty` | URL เว็บที่ deploy แล้ว (optional) |
| `source_url` | `string` | `source_url,omitempty` | URL source code (optional) |
| `sort_order` | `int` | `sort_order` | ลำดับการแสดงผล |
| `created_at` | `time.Time` | `created_at` | วันที่สร้าง |
| `updated_at` | `time.Time` | `updated_at` | วันที่แก้ไขล่าสุด |

---

## 7. experiences

ประสบการณ์การทำงาน — หลาย documents, เรียงตาม `sort_order`

### Schema

| Field | Type | BSON Tag | คำอธิบาย |
|-------|------|----------|---------|
| `_id` | `ObjectID` | `_id,omitempty` | Primary key |
| `role` | `string` | `role` | ตำแหน่งงาน |
| `company` | `string` | `company` | ชื่อบริษัท |
| `period` | `string` | `period` | ช่วงเวลา เช่น "2024 - Present" |
| `description` | `string` | `description` | คำอธิบายงาน |
| `highlights` | `[]string` | `highlights` | จุดเด่นของงาน |
| `sort_order` | `int` | `sort_order` | ลำดับการแสดงผล |
| `created_at` | `time.Time` | `created_at` | วันที่สร้าง |
| `updated_at` | `time.Time` | `updated_at` | วันที่แก้ไขล่าสุด |

---

## 8. social_links

ลิงก์โซเชียลมีเดีย — หลาย documents, เรียงตาม `sort_order`

### Schema

| Field | Type | BSON Tag | คำอธิบาย |
|-------|------|----------|---------|
| `_id` | `ObjectID` | `_id,omitempty` | Primary key |
| `name` | `string` | `name` | ชื่อ platform เช่น "GitHub" |
| `url` | `string` | `url` | URL ลิงก์ |
| `icon` | `string` | `icon` | ชื่อ icon (Iconify) เช่น "mdi:github" |
| `sort_order` | `int` | `sort_order` | ลำดับการแสดงผล |
| `created_at` | `time.Time` | `created_at` | วันที่สร้าง |
| `updated_at` | `time.Time` | `updated_at` | วันที่แก้ไขล่าสุด |

---

## 9. contact_messages

ข้อความที่ผู้เยี่ยมชมส่งมาผ่าน contact form

### Schema

| Field | Type | BSON Tag | คำอธิบาย |
|-------|------|----------|---------|
| `_id` | `ObjectID` | `_id,omitempty` | Primary key |
| `name` | `string` | `name` | ชื่อผู้ส่ง |
| `email` | `string` | `email` | อีเมลผู้ส่ง |
| `subject` | `string` | `subject` | หัวข้อ |
| `message` | `string` | `message` | เนื้อหาข้อความ |
| `is_read` | `bool` | `is_read` | อ่านแล้วหรือยัง (default: false) |
| `created_at` | `time.Time` | `created_at` | วันที่ส่ง |

---

## ความสัมพันธ์ระหว่าง Collections

```
                    ┌─────────────────┐
                    │  admin_users    │
                    │  (auth + RBAC)  │
                    └────────┬────────┘
                             │ JWT token contains
                             │ user_id + role
                             ▼
        ┌────────────────────────────────────────┐
        │            Admin Operations            │
        │  (CRUD via JWT-authenticated requests) │
        └────┬───────┬───────┬──────┬───────┬────┘
             │       │       │      │       │
             ▼       ▼       ▼      ▼       ▼
     ┌───────┐ ┌─────┐ ┌─────┐ ┌──────┐ ┌──────────┐
     │skills │ │hero │ │about│ │sites │ │projects  │
     └───────┘ └─────┘ └─────┘ │setts │ └──────────┘
                               └──────┘
             │       │       │
             ▼       ▼       ▼
     ┌───────────┐ ┌──────────┐ ┌────────────────┐
     │experiences│ │social    │ │contact         │
     │           │ │links     │ │messages        │
     └───────────┘ └──────────┘ └────────────────┘
                                       ▲
                                       │ POST /contact
                                       │ (public, no auth)
                                ┌──────┴──────┐
                                │  Website    │
                                │  Visitor    │
                                └─────────────┘
```

**หมายเหตุ:**
- ไม่มี foreign key references ระหว่าง collections (MongoDB ไม่ enforce foreign keys)
- `admin_users` เชื่อมกับ operations อื่น ๆ ผ่าน JWT token (ไม่ได้ join ใน DB)
- Singleton collections (site_settings, hero, about) ใช้ Upsert pattern — ถ้ายังไม่มี document จะสร้างใหม่

---

## Seed Data (ข้อมูลเริ่มต้น)

เมื่อรัน server ครั้งแรกและ `admin_users` collection ว่าง ระบบจะ seed ข้อมูลต่อไปนี้:

| Collection | จำนวน Documents | คำอธิบาย |
|-----------|----------------|---------|
| `admin_users` | 1 | Admin user (username/password จาก env vars) |
| `site_settings` | 1 | Default site settings (title, theme, meta) |
| `hero` | 1 | Default hero content |
| `about` | 1 | Default about content พร้อม stats 4 รายการ |
| `skills` | 14 | ตัวอย่าง skills (Vue, Go, Docker ฯลฯ) |
| `projects` | 2 | ตัวอย่าง projects |
| `experiences` | 2 | ตัวอย่าง experiences |
| `social_links` | 4 | ตัวอย่าง links (GitHub, LinkedIn, Twitter, Email) |

---

## Sort Order Convention

Collections ที่รองรับ Reorder API (`PUT .../reorder`):
- `projects` — `PUT /api/v1/admin/projects/reorder`
- `experiences` — `PUT /api/v1/admin/experiences/reorder`
- `social_links` — `PUT /api/v1/admin/social-links/reorder`

**⚠ skills ไม่มี reorder endpoint** — `sort_order` ของ skills กำหนดตอน `POST` (สร้าง) เท่านั้น และไม่สามารถเปลี่ยนผ่าน `PUT` (update จะไม่แก้ `sort_order`)

**วิธีการทำงาน:**
- `sort_order` เป็น `int` เริ่มจาก `0`
- `List` query จะเรียงตาม `sort_order` ascending
- Reorder endpoint รับ array ของ IDs → กำหนด sort_order ตาม index position

```json
// PUT /api/v1/admin/projects/reorder
{
  "ids": ["id_c", "id_a", "id_b"]
}
// ผลลัพธ์: id_c → sort_order: 0, id_a → sort_order: 1, id_b → sort_order: 2
```

---

## Default Sort Order ของ List Queries

| Collection | Sort Field | Direction | คำอธิบาย |
|-----------|-----------|-----------|---------|
| `skills` | `sort_order` | Ascending (1) | เรียงตามลำดับที่กำหนด |
| `projects` | `sort_order` | Ascending (1) | เรียงตามลำดับที่กำหนด |
| `experiences` | `sort_order` | Ascending (1) | เรียงตามลำดับที่กำหนด |
| `social_links` | `sort_order` | Ascending (1) | เรียงตามลำดับที่กำหนด |
| `contact_messages` | `created_at` | **Descending (-1)** | ข้อความล่าสุดขึ้นก่อน |
| `admin_users` | `created_at` | Ascending (1) | สร้างก่อนขึ้นก่อน |

---

## MongoDB Connection Settings

| Parameter | ค่า | คำอธิบาย |
|-----------|-----|---------|
| Connection URI | `MONGO_URI` env var | เช่น `mongodb://localhost:27017` |
| Database | `MONGO_DB` env var | เช่น `portfolio_admin` |
| Connect timeout | 10 seconds | รอเชื่อมต่อ |
| Disconnect timeout | 5 seconds | รอปิดการเชื่อมต่อ |
| Read preference | Primary | อ่านจาก primary node |
