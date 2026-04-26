# Setup & Deployment Guide — Portfolio Admin API

เอกสารนี้อธิบายวิธีตั้งค่าโปรเจกต์ตั้งแต่เริ่มต้น ทั้งแบบ Local Development และ Docker พร้อม troubleshooting

---

## สิ่งที่ต้องติดตั้งก่อน (Prerequisites)

### สำหรับรันแบบ Docker (แนะนำ)

| Software | Version | ดาวน์โหลด |
|---------|---------|----------|
| Docker Desktop | ล่าสุด | [docker.com](https://www.docker.com/products/docker-desktop/) |
| Git | ล่าสุด | [git-scm.com](https://git-scm.com/) |

### สำหรับรันแบบ Local (ไม่ใช้ Docker)

| Software | Version | ดาวน์โหลด |
|---------|---------|----------|
| Go | 1.22+ | [go.dev/dl](https://go.dev/dl/) |
| MongoDB | 7+ | [mongodb.com](https://www.mongodb.com/try/download/community) |
| Git | ล่าสุด | [git-scm.com](https://git-scm.com/) |

---

## วิธีที่ 1: รันด้วย Docker Compose (แนะนำ)

### Step 1: Clone โปรเจกต์

```bash
git clone <repository-url>
cd Admin_Website_Management
```

### Step 2: สร้างไฟล์ .env

```bash
# Copy จาก .env.example
cp .env.example .env
```

แก้ไขค่าใน `.env` ตามต้องการ:

```env
# Server
PORT=8080

# MongoDB
MONGO_URI=mongodb://mongo:27017
MONGO_DB=portfolio_admin

# JWT — ⚠️ ต้องเปลี่ยนใน production!
JWT_SECRET=your-strong-random-secret-key-here

# Initial Admin (ใช้ตอน seed ครั้งแรก)
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-secure-password

# File Uploads
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=10

# CORS — ใส่ origin ของ frontend (หลายโดเมน คั่นด้วย comma — รองรับหลายไซต์ / admin แยกโดเมน)
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,https://admin.example.com
```

**ข้อมูล Seed เริ่มต้น:** ครั้งแรกที่ DB ว่าง ระบบจะ seed ให้อัตโนมัติ — สร้างผู้ใช้ admin, site เริ่มต้น (ชื่อ `My Portfolio`, slug `default`, type `portfolio`, domains `["localhost:3000"]`), แถว `site_member` (admin เป็น owner ของ site นี้) และข้อมูล portfolio ทั้งชุดที่ผูก `site_id` กับ site เริ่มต้น

**เส้นทาง API แบบ multi-site (เทียบของเดิม):**

| เดิม | ใหม่ |
|------|------|
| `GET /api/v1/portfolio` | `GET /api/v1/public/sites/{siteId}/portfolio` |
| `GET/POST /api/v1/admin/skills` ฯลฯ | `.../api/v1/admin/sites/{siteId}/portfolio/skills` |
| ส่วนอื่นของ portfolio (hero, about, projects, experiences, social-links, contacts, upload, site-settings) | อยู่ภายใต้ `/api/v1/admin/sites/{siteId}/portfolio/...` ตามชื่อ resource |
| Submit contact สาธารณะ | `POST /api/v1/public/sites/{siteId}/portfolio/contacts` |

### Step 3: รัน Docker Compose

```bash
# Build และรัน (ครั้งแรก)
docker compose up --build

# รันแบบ background (detach mode)
docker compose up --build -d

# ดู logs
docker compose logs -f api

# หยุด services
docker compose down

# หยุดและลบ volumes (⚠️ ลบข้อมูล DB ทั้งหมด)
docker compose down -v
```

### Step 4: ทดสอบ

```bash
# แก้ไข public routes ให้มี siteId — ดึง site จาก domain ที่ seed ไว้ (localhost:3000)
curl "http://localhost:8080/api/v1/public/sites/by-domain?host=localhost%3A3000"

# ใช้ค่า _id จาก JSON ด้านบน แทน {siteId} แล้วดึง portfolio สาธารณะ
curl "http://localhost:8080/api/v1/public/sites/{siteId}/portfolio"

# ตัวอย่าง admin (ต้องมี JWT หลัง login) — CRUD skills ของไซต์นั้น
# curl -H "Authorization: Bearer <token>" "http://localhost:8080/api/v1/admin/sites/{siteId}/portfolio/skills"

# ทดสอบ login
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-secure-password"}'
```

---

## วิธีที่ 2: รันแบบ Local Development

### Step 1: Clone โปรเจกต์

```bash
git clone <repository-url>
cd Admin_Website_Management
```

### Step 2: ติดตั้ง Go dependencies

```bash
go mod tidy
```

### Step 3: รัน MongoDB

**ตัวเลือก A: MongoDB ติดตั้งในเครื่อง**
```bash
# macOS (Homebrew)
brew services start mongodb-community

# Windows — เปิด MongoDB service จาก Services manager
# หรือรัน mongod โดยตรง
mongod --dbpath /data/db
```

**ตัวเลือก B: MongoDB ผ่าน Docker (ไม่ต้องติดตั้ง)**
```bash
docker run -d \
  --name portfolio-mongo \
  -p 27017:27017 \
  -v portfolio_mongo_data:/data/db \
  mongo:7
```

### Step 4: สร้างไฟล์ .env

```bash
cp .env.example .env
```

แก้ไข `MONGO_URI` ให้ชี้ไป localhost (และตั้ง CORS ให้ครบทุก origin ที่ใช้ — รองรับหลายโดเมนสำหรับ multi-site):

```env
MONGO_URI=mongodb://localhost:27017
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,https://admin.example.com
```

### Step 5: รัน Server

```bash
go run ./cmd/server
```

Server จะเริ่มทำงานที่ `http://localhost:8080`

**ตัวอย่าง output:**
```json
{"time":"2025-01-15T10:00:00Z","level":"INFO","msg":"connected to MongoDB","database":"portfolio_admin"}
{"time":"2025-01-15T10:00:00Z","level":"INFO","msg":"seeding database with initial data"}
{"time":"2025-01-15T10:00:00Z","level":"INFO","msg":"database seeded successfully"}
{"time":"2025-01-15T10:00:00Z","level":"INFO","msg":"server starting","port":"8080"}
```

### Step 6: Build Binary (Optional)

```bash
# Build สำหรับ OS ปัจจุบัน
go build -o bin/server ./cmd/server

# รัน binary
./bin/server

# Cross-compile สำหรับ Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/server-linux ./cmd/server
```

---

## Environment Variables Reference

| Variable | Required | Default | คำอธิบาย |
|----------|----------|---------|---------|
| `PORT` | No | `8080` | พอร์ตที่ API server listen |
| `MONGO_URI` | No | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGO_DB` | No | `portfolio_admin` | ชื่อ database |
| `JWT_SECRET` | **Yes** (production) | `change-me-in-production` | Secret key สำหรับ sign JWT token |
| `ADMIN_USERNAME` | No | `admin` | Username ของ admin ที่ seed ตอนเริ่มต้น |
| `ADMIN_PASSWORD` | No | `changeme123` | Password ของ admin ที่ seed ตอนเริ่มต้น |
| `UPLOAD_DIR` | No | `./uploads` | Directory สำหรับเก็บไฟล์ที่อัปโหลด |
| `MAX_UPLOAD_SIZE_MB` | No | `10` | ขนาดไฟล์สูงสุดที่อัปโหลดได้ (MB) |
| `ALLOWED_ORIGINS` | No | `http://localhost:3000` | CORS allowed origins หลายค่าได้ (คั่นด้วย `,`) — ใช้เมื่อมีหลายโดเมน (public site / admin ฯลฯ) |

### สิ่งที่ต้องเปลี่ยนใน Production

1. **`JWT_SECRET`** — ต้องเปลี่ยนเป็น random string ยาว ๆ (เช่น 64+ characters)
2. **`ADMIN_PASSWORD`** — ต้องเปลี่ยนเป็นรหัสผ่านที่ปลอดภัย
3. **`ALLOWED_ORIGINS`** — ต้องใส่ domain จริงของทุก frontend ที่เรียก API (หลายค่าคั่นด้วย comma) เพื่อรองรับหลายไซต์
4. **`MONGO_URI`** — ใช้ connection string จริงพร้อม authentication

---

## Project Structure สำหรับ Development

```
Admin_Website_Management/
├── cmd/server/main.go       ← ★ Entry point — เริ่มอ่านที่นี่
├── internal/                ← Business logic (private packages)
│   ├── config/              ← อ่าน env vars
│   ├── database/            ← MongoDB connection + seed
│   ├── handler/             ← HTTP request handlers
│   ├── middleware/          ← Middleware (CORS, Auth, Logging)
│   ├── model/               ← Data models + DTOs
│   ├── repository/          ← Database operations
│   └── router/              ← Route registration
├── pkg/response/            ← JSON response helpers (public package)
├── GUIDE/                   ← เอกสารสำหรับนักพัฒนา
├── Dockerfile               ← Docker build config
├── docker-compose.yml       ← Docker orchestration
├── go.mod                   ← Go dependencies
└── .env.example             ← Template environment variables
```

### ลำดับการอ่านโค้ดสำหรับคนใหม่

1. **`cmd/server/main.go`** — เข้าใจ startup flow ทั้งหมด
2. **`internal/config/config.go`** — ดูว่ามี config อะไรบ้าง
3. **`internal/model/model.go`** — เข้าใจโครงสร้างข้อมูล
4. **`internal/router/router.go`** — ดู routes ทั้งหมดและ RBAC
5. **`internal/middleware/middleware.go`** — เข้าใจ middleware stack
6. **`internal/handler/public.go`** — ดูตัวอย่าง handler ง่าย ๆ
7. **`internal/handler/auth.go`** — เข้าใจ authentication flow
8. **`internal/repository/`** — ดู database operations

---

## Docker Commands ที่ใช้บ่อย

```bash
# สร้างและรัน services
docker compose up --build -d

# ดู status ของ services
docker compose ps

# ดู logs ของ API
docker compose logs -f api

# ดู logs ของ MongoDB
docker compose logs -f mongo

# รีสตาร์ท API (หลังแก้โค้ด)
docker compose up --build -d api

# เข้าไปใน container
docker compose exec api sh
docker compose exec mongo mongosh

# ดูข้อมูลใน MongoDB
docker compose exec mongo mongosh portfolio_admin --eval "db.admin_users.find().pretty()"

# ลบทุกอย่างและเริ่มใหม่
docker compose down -v
docker compose up --build -d
```

---

## Useful MongoDB Commands

```javascript
// เชื่อมต่อ MongoDB
mongosh mongodb://localhost:27017/portfolio_admin

// ดู collections ทั้งหมด
show collections

// ดู admin users
db.admin_users.find().pretty()

// ดู projects
db.projects.find().sort({ sort_order: 1 }).pretty()

// ดู contact messages (ยังไม่อ่าน)
db.contact_messages.find({ is_read: false }).pretty()

// นับ documents
db.skills.countDocuments()

// ลบ database ทั้งหมด (⚠️ ระวัง!)
db.dropDatabase()
```

---

## Healthcheck

Public portfolio เดิมที่ `/api/v1/portfolio` ถูกย้ายเป็นแบบ multi-site แล้ว — ตรวจสุขภาพควรใช้ endpoint ที่ไม่ต้องรู้ `siteId` ล่วงหน้า เช่น resolve โดเมน (หลัง seed จะมี `localhost:3000`):

```bash
# ตรวจสอบว่า API + DB + seed ทำงานสอดคล้องกัน (ควรได้ 200 และ JSON ของ site)
curl "http://localhost:8080/api/v1/public/sites/by-domain?host=localhost%3A3000"

# หรือถ้ามี siteId แล้ว จะยืนยัน portfolio ได้โดยตรง
# curl "http://localhost:8080/api/v1/public/sites/{siteId}/portfolio"
```

**หมายเหตุ:** ถ้า `Dockerfile` ยังอ้าง `HEALTHCHECK` ไปที่ `/api/v1/portfolio` ควรอัปเดตเป็น URL ที่เหมาะสม (เช่น `by-domain` ด้านบน) ให้สอดคล้องกับสถาปัตยกรรมใหม่

Docker Compose จะ:
- ตรวจสอบ API ทุก 30 วินาที
- ตรวจสอบ MongoDB ทุก 10 วินาที
- ถ้า unhealthy → จะ restart container อัตโนมัติ

---

## Troubleshooting

### 1. "failed to connect to database"

**สาเหตุ:** MongoDB ยังไม่พร้อมหรือ connection string ผิด

**แก้ไข:**
```bash
# ตรวจสอบ MongoDB ทำงาน
docker compose ps mongo

# ตรวจสอบ logs
docker compose logs mongo

# ถ้ารัน local: ตรวจสอบว่า MongoDB service เริ่มแล้ว
mongosh --eval "db.adminCommand('ping')"
```

### 2. "invalid credentials" ตอน Login

**สาเหตุ:** Username/password ไม่ตรงกับที่ seed ไว้

**แก้ไข:**
```bash
# ตรวจสอบค่าใน .env
cat .env | grep ADMIN

# ตรวจสอบ user ใน DB
docker compose exec mongo mongosh portfolio_admin --eval "db.admin_users.find({}, {username:1, role:1})"

# ถ้าต้องการ reset → ลบ admin_users แล้ว restart (จะ seed ใหม่)
docker compose exec mongo mongosh portfolio_admin --eval "db.admin_users.drop()"
docker compose restart api
```

### 3. "CORS error" จาก Frontend

**สาเหตุ:** Origin ของ frontend ไม่อยู่ใน `ALLOWED_ORIGINS`

**แก้ไข:**
```bash
# เพิ่ม origin ใน .env (หลายโดเมนสำหรับหลายไซต์ / แอดมิน)
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,https://admin.example.com,https://your-public-site.com

# Restart API
docker compose restart api
```

### 4. "file too large" ตอน Upload

**สาเหตุ:** ไฟล์ใหญ่กว่า `MAX_UPLOAD_SIZE_MB`

**แก้ไข:**
```bash
# เพิ่มขนาดใน .env
MAX_UPLOAD_SIZE_MB=20

# Restart API
docker compose restart api
```

### 5. Port 8080 ถูกใช้งานอยู่

**แก้ไข:**
```bash
# เปลี่ยน port ใน .env
PORT=9090

# แก้ docker-compose.yml ports mapping ด้วย
# ports: "9090:9090"

# หรือหา process ที่ใช้ port 8080
# Windows
netstat -ano | findstr :8080
# macOS/Linux
lsof -i :8080
```

### 6. ข้อมูลหายหลัง Docker Compose Down

**สาเหตุ:** ใช้ `docker compose down -v` (flag `-v` ลบ volumes)

**ป้องกัน:**
```bash
# ใช้ down โดยไม่มี -v flag
docker compose down

# ตรวจสอบ volumes
docker volume ls | grep portfolio
```

---

## Production Deployment Checklist

- [ ] เปลี่ยน `JWT_SECRET` เป็น random string ที่แข็งแกร่ง (64+ characters)
- [ ] เปลี่ยน `ADMIN_PASSWORD` เป็นรหัสผ่านที่ปลอดภัย
- [ ] ตั้งค่า `ALLOWED_ORIGINS` ให้ครบทุก domain จริงที่เรียก API (คั่นด้วย comma)
- [ ] ใช้ MongoDB ที่มี authentication (`MONGO_URI` พร้อม username/password)
- [ ] ตั้งค่า HTTPS (TLS) ผ่าน reverse proxy (nginx/traefik)
- [ ] ตั้งค่า persistent storage สำหรับ uploads volume
- [ ] ตั้งค่า MongoDB backup strategy
- [ ] ตั้งค่า monitoring/alerting สำหรับ API health
- [ ] ตรวจสอบว่า `.env` ไม่ถูก commit ลง git
- [ ] ตั้งค่า log aggregation (เช่น Loki, ELK)
