# RBAC Migration Guide

คู่มือสำหรับปรับ **เว็บหลังบ้าน (Admin Panel)** และ **เว็บหน้าบ้าน (Portfolio)** ให้รองรับระบบ Role-Based Access Control (RBAC) ที่เพิ่มใน Backend

---

## สรุป Role ทั้ง 3 ระดับ

| Role | ระดับ | สิทธิ์ |
|------|-------|--------|
| `admin` | 3 | เข้าถึงทุกอย่าง + จัดการ user (เพิ่ม/ลบ/เปลี่ยนรหัส/กำหนด role) |
| `user_account` | 2 | CRUD content ทุกอย่าง + ลบ contact messages + upload ไฟล์ |
| `visitor` | 1 | อ่าน contact messages ได้อย่างเดียว (list + get) ลบไม่ได้ แก้ไข content ไม่ได้ |

---

## API ที่เปลี่ยนแปลง

### Login Response (เปลี่ยน)

**ก่อน:**
```json
{
  "data": {
    "token": "eyJ..."
  }
}
```

**หลัง:**
```json
{
  "data": {
    "token": "eyJ...",
    "user": {
      "id": "682...",
      "username": "admin",
      "role": "admin"
    }
  }
}
```

### API ใหม่ที่เพิ่ม

| Method | Endpoint | Role ที่ต้องการ | คำอธิบาย |
|--------|----------|----------------|----------|
| `GET` | `/api/v1/admin/auth/me` | visitor+ | ดึงข้อมูล user ปัจจุบัน |
| `GET` | `/api/v1/admin/users` | admin | ดึงรายชื่อ user ทั้งหมด |
| `POST` | `/api/v1/admin/users` | admin | สร้าง user ใหม่ |
| `GET` | `/api/v1/admin/users/{id}` | admin | ดึงข้อมูล user ตาม ID |
| `PUT` | `/api/v1/admin/users/{id}` | admin | แก้ไข username/role |
| `DELETE` | `/api/v1/admin/users/{id}` | admin | ลบ user |
| `PUT` | `/api/v1/admin/users/{id}/password` | admin | เปลี่ยนรหัสผ่าน user |

### API เดิมที่เปลี่ยน Role Requirement

| Endpoint | ก่อน (แค่ login) | หลัง (ต้องมี role) |
|----------|------------------|---------------------|
| Content CRUD (site-settings, hero, about, skills, projects, experiences, social-links) | auth only | `user_account`+ |
| `GET /admin/contacts`, `GET /admin/contacts/{id}` | auth only | `visitor`+ |
| `DELETE /admin/contacts/{id}` | auth only | `user_account`+ |
| `POST /admin/upload` | auth only | `user_account`+ |

### Error Response ใหม่

เมื่อ role ไม่เพียงพอจะได้ HTTP 403:
```json
{
  "error": "insufficient permissions"
}
```

---

## Request/Response สำหรับ User Management API

### POST /api/v1/admin/users (สร้าง user)

**Request:**
```json
{
  "username": "editor01",
  "password": "SecurePass123!",
  "role": "user_account"
}
```
- `role` ต้องเป็น: `admin`, `user_account`, หรือ `visitor`
- `password` ต้องยาวอย่างน้อย 8 ตัวอักษร

**Response (201):**
```json
{
  "data": {
    "id": "682...",
    "username": "editor01",
    "role": "user_account",
    "created_at": "2026-04-11T...",
    "updated_at": "2026-04-11T..."
  }
}
```

### PUT /api/v1/admin/users/{id} (แก้ไข user)

**Request:**
```json
{
  "username": "editor01_updated",
  "role": "visitor"
}
```

### PUT /api/v1/admin/users/{id}/password (เปลี่ยนรหัสผ่าน)

**Request:**
```json
{
  "new_password": "NewSecurePass456!"
}
```

### Safety Rules ของ API

- ลบ admin คนสุดท้ายไม่ได้ → `400: cannot delete the last admin`
- ลบตัวเองไม่ได้ → `400: cannot delete your own account`
- Downgrade role ตัวเองไม่ได้ → `400: cannot downgrade your own role`
- Username ซ้ำไม่ได้ → `409: username already exists`

---

## สิ่งที่ต้องปรับในเว็บหลังบ้าน (Admin Panel)

### 1. แก้ไข Login Flow

ตอน login สำเร็จ ต้องเก็บ `user.role` ไว้ด้วย (ไม่ใช่แค่ token):

```typescript
const login = async (username: string, password: string) => {
  const { data } = await $fetch('/api/v1/admin/auth/login', {
    method: 'POST',
    body: { username, password },
  })
  
  // เก็บ token
  authToken.value = data.token
  
  // เก็บ user info + role (ใหม่)
  currentUser.value = data.user  // { id, username, role }
}
```

### 2. สร้าง Composable สำหรับ Permission Check

สร้าง `composables/useAuth.ts` (หรือแก้ไขถ้ามีอยู่แล้ว):

```typescript
export const useAuth = () => {
  const user = useState<{ id: string; username: string; role: string } | null>('auth-user', () => null)
  const token = useState<string | null>('auth-token', () => null)

  const isAdmin = computed(() => user.value?.role === 'admin')
  const isUserAccount = computed(() => 
    user.value?.role === 'admin' || user.value?.role === 'user_account'
  )
  const isVisitor = computed(() => !!user.value?.role)

  const hasRole = (minRole: string) => {
    const levels: Record<string, number> = { visitor: 1, user_account: 2, admin: 3 }
    const userLevel = levels[user.value?.role ?? ''] ?? 0
    return userLevel >= (levels[minRole] ?? 99)
  }

  return { user, token, isAdmin, isUserAccount, isVisitor, hasRole }
}
```

### 3. ซ่อน/แสดง เมนูตาม Role

ใน sidebar หรือ navigation ของ admin panel:

```vue
<template>
  <!-- เมนูที่ทุก role เห็น -->
  <NavItem to="/admin/contacts" label="Contact Messages" />

  <!-- เมนูที่ user_account+ เห็น -->
  <template v-if="isUserAccount">
    <NavItem to="/admin/hero" label="Hero" />
    <NavItem to="/admin/about" label="About" />
    <NavItem to="/admin/skills" label="Skills" />
    <NavItem to="/admin/projects" label="Projects" />
    <NavItem to="/admin/experiences" label="Experiences" />
    <NavItem to="/admin/social-links" label="Social Links" />
    <NavItem to="/admin/site-settings" label="Site Settings" />
  </template>

  <!-- เมนูที่ admin เท่านั้นเห็น -->
  <NavItem v-if="isAdmin" to="/admin/users" label="User Management" />
</template>

<script setup lang="ts">
const { isAdmin, isUserAccount } = useAuth()
</script>
```

### 4. สร้างหน้า User Management (admin only)

สร้างหน้าใหม่ `pages/admin/users/index.vue`:
- แสดงตาราง user ทั้งหมด (GET /api/v1/admin/users)
- ปุ่มเพิ่ม user ใหม่
- ปุ่มแก้ไข / ลบ / เปลี่ยนรหัสผ่าน ในแต่ละแถว

สร้าง dialog/modal สำหรับ:
- **Create User**: form username + password + role (dropdown)
- **Edit User**: form username + role
- **Change Password**: form new_password
- **Delete Confirmation**: ยืนยันก่อนลบ

### 5. ป้องกัน Route ด้วย Middleware

สร้าง Nuxt middleware เพื่อเช็ค role ก่อนเข้าหน้า:

```typescript
// middleware/role.ts
export default defineNuxtRouteMiddleware((to) => {
  const { hasRole } = useAuth()
  
  const roleMap: Record<string, string> = {
    '/admin/users': 'admin',
    '/admin/hero': 'user_account',
    '/admin/about': 'user_account',
    '/admin/skills': 'user_account',
    '/admin/projects': 'user_account',
    '/admin/experiences': 'user_account',
    '/admin/social-links': 'user_account',
    '/admin/site-settings': 'user_account',
    '/admin/contacts': 'visitor',
  }

  for (const [path, role] of Object.entries(roleMap)) {
    if (to.path.startsWith(path) && !hasRole(role)) {
      return navigateTo('/admin/contacts')
    }
  }
})
```

### 6. ซ่อนปุ่มลบใน Contact Messages สำหรับ Visitor

ในหน้า contact messages ให้เช็ค role ก่อนแสดงปุ่มลบ:

```vue
<button
  v-if="isUserAccount"
  data-testid="delete-message"
  @click="handleDelete(message.id)"
>
  Delete
</button>
```

### 7. Handle HTTP 403 Error

เพิ่ม error handling สำหรับ 403 Forbidden ใน API interceptor:

```typescript
if (response.status === 403) {
  // แสดง toast/notification
  showNotification('คุณไม่มีสิทธิ์เข้าถึงส่วนนี้', 'error')
}
```

---

## สิ่งที่ต้องปรับในเว็บหน้าบ้าน (Portfolio)

### ไม่ต้องปรับอะไร

เว็บหน้าบ้านใช้แค่ 2 endpoints ที่เป็น Public (ไม่ต้อง login):

| Endpoint | สถานะ |
|----------|--------|
| `GET /api/v1/portfolio` | ไม่เปลี่ยน — Public, ไม่ต้อง auth |
| `POST /api/v1/contact` | ไม่เปลี่ยน — Public, ไม่ต้อง auth |

ทั้ง 2 endpoints ไม่ได้รับผลกระทบจากการเพิ่ม RBAC เพราะไม่ต้องผ่าน Auth middleware

---

## Database Migration

### สำหรับ Database ที่มีอยู่แล้ว (มี user เก่าที่ยังไม่มี role)

ถ้า MongoDB มี admin_users collection อยู่แล้ว ต้อง update ให้มี `role` field:

```javascript
// รันใน mongosh
use portfolio_admin

// เพิ่ม role ให้ user เดิมทั้งหมด
db.admin_users.updateMany(
  { role: { $exists: false } },
  { $set: { role: "admin", updated_at: new Date() } }
)

// ยืนยัน
db.admin_users.find({}, { username: 1, role: 1 })
```

### ทางเลือก: Reset แล้ว Seed ใหม่

ถ้าอยู่ในช่วง development สามารถลบ collection แล้วให้ seed ใหม่:

```javascript
use portfolio_admin
db.admin_users.drop()
```

แล้ว restart API server — seed จะสร้าง admin user ใหม่พร้อม `role: "admin"`

---

## JWT Token Changes

JWT claims เพิ่ม field `role`:

**ก่อน:**
```json
{
  "sub": "user_id",
  "usr": "admin",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**หลัง:**
```json
{
  "sub": "user_id",
  "usr": "admin",
  "role": "admin",
  "exp": 1234567890,
  "iat": 1234567890
}
```

> **สำคัญ:** Token เก่าที่ไม่มี `role` จะยังใช้ login ได้ แต่จะไม่มีสิทธิ์เข้าถึง endpoint ที่ต้องการ role (middleware จะเห็น role เป็น empty string → level 0) ดังนั้นหลัง deploy RBAC แล้ว **user ต้อง login ใหม่** เพื่อรับ token ที่มี role

---

## Checklist

### Backend (โปรเจกต์นี้) — เสร็จแล้ว
- [x] เพิ่ม `role` field ใน AdminUser model
- [x] เพิ่ม CRUD methods ใน AdminUserRepository
- [x] เพิ่ม RequireRole middleware
- [x] สร้าง UserHandler (CRUD + change password)
- [x] เพิ่ม `role` ใน JWT claims + Login response
- [x] แยก route ตาม role (admin / user_account / visitor)
- [x] Seed admin user พร้อม `role: "admin"`
- [x] Unique index บน username

### เว็บหลังบ้าน (Admin Panel) — ต้องทำ
- [ ] แก้ login flow เพื่อเก็บ role
- [ ] สร้าง useAuth composable พร้อม permission check
- [ ] ซ่อน/แสดงเมนูตาม role
- [ ] สร้างหน้า User Management
- [ ] เพิ่ม route middleware เช็ค role
- [ ] ซ่อนปุ่มลบ contact สำหรับ visitor
- [ ] Handle HTTP 403 error

### เว็บหน้าบ้าน (Portfolio) — ไม่ต้องปรับ
- [x] ไม่มีผลกระทบ (ใช้แค่ Public API)
