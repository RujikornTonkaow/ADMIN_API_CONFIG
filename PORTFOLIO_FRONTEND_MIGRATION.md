# Portfolio Frontend — Migration Guide

เอกสารนี้อธิบายสิ่งที่ต้องเปลี่ยนแปลงในโปรเจกต์ **Portfolio Frontend** เพื่อดึงข้อมูลจาก **Admin Panel API** แทนการ hardcode

---

## 1. ภาพรวมการเปลี่ยนแปลง

| สิ่งที่ต้องทำ | รายละเอียด |
|---------------|-----------|
| เพิ่ม API Base URL config | กำหนดค่า URL ของ Admin API ใน `runtimeConfig` |
| แก้ไข `usePortfolioData.ts` | เปลี่ยนจาก hardcode เป็น fetch จาก API |
| แก้ไข `SectionContact.vue` | เพิ่มการ POST ข้อมูล contact form ไปยัง API |
| แก้ไข `SectionHero.vue` | ดึง greeting, full_name, subtitle, CTA จาก API data |
| แก้ไข `SectionAbout.vue` | ดึง title, bio, tags, stats จาก API data |
| แก้ไข `TheNavbar.vue` | ดึง nav items และ site title จาก API data |
| แก้ไข `TheFooter.vue` | ดึง brand name และ tagline จาก API data |
| แก้ไข `useTheme.ts` | ใช้ `default_theme` จาก API แทน hardcode |
| แก้ไข `nuxt.config.ts` | เพิ่ม `runtimeConfig` สำหรับ API URL |
| แก้ไข `pages/index.vue` | ใช้ SEO metadata จาก API data |

---

## 2. เพิ่ม API Configuration

### 2.1 `nuxt.config.ts`

```typescript
export default defineNuxtConfig({
  runtimeConfig: {
    public: {
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'
    }
  },
  // ... existing config
})
```

### 2.2 สร้างไฟล์ `.env`

```env
NUXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

---

## 3. แก้ไข `composables/usePortfolioData.ts`

**ลบ** ข้อมูล hardcode ทั้งหมด แล้วแทนที่ด้วย:

```typescript
interface PortfolioData {
  site_settings: {
    id: string
    site_title: string
    page_title: string
    meta_description: string
    footer_tagline: string
    default_theme: 'midnight' | 'sunshine'
    profile_image: string
  }
  hero: {
    greeting: string
    full_name: string
    subtitle: string
    cta_primary_text: string
    cta_primary_link: string
    cta_secondary_text: string
    cta_secondary_link: string
  }
  about: {
    title: string
    bio_paragraphs: string[]
    personality_tags: string[]
    stats: { value: string; label: string }[]
  }
  skills: {
    id: string
    name: string
    icon: string
    category: 'frontend' | 'backend' | 'devops' | 'tools'
    sort_order: number
  }[]
  projects: {
    id: string
    title: string
    description: string
    tags: string[]
    image?: string
    live_url?: string
    source_url?: string
    sort_order: number
  }[]
  experiences: {
    id: string
    role: string
    company: string
    period: string
    description: string
    highlights: string[]
    sort_order: number
  }[]
  social_links: {
    id: string
    name: string
    url: string
    icon: string
    sort_order: number
  }[]
  nav_items: {
    label: string
    href: string
  }[]
}

export const usePortfolioData = () => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBaseUrl

  const { data, pending, error } = useFetch<{ data: PortfolioData }>(
    `${apiBase}/api/v1/portfolio`,
    {
      key: 'portfolio-data',
      default: () => ({ data: null }),
    }
  )

  const portfolio = computed(() => data.value?.data ?? null)

  const siteSettings = computed(() => portfolio.value?.site_settings)
  const hero = computed(() => portfolio.value?.hero)
  const about = computed(() => portfolio.value?.about)
  const skills = computed(() => portfolio.value?.skills ?? [])
  const projects = computed(() => portfolio.value?.projects ?? [])
  const experiences = computed(() => portfolio.value?.experiences ?? [])
  const socialLinks = computed(() => portfolio.value?.social_links ?? [])
  const navItems = computed(() => portfolio.value?.nav_items ?? [])

  return {
    portfolio,
    siteSettings,
    hero,
    about,
    skills,
    projects,
    experiences,
    socialLinks,
    navItems,
    pending,
    error,
  }
}
```

---

## 4. แก้ไข Components

### 4.1 `SectionHero.vue`

**เปลี่ยนจาก:**
- ข้อความ greeting, ชื่อ, subtitle ที่ hardcode ใน `<template>`
- ปุ่ม CTA ที่ hardcode

**เปลี่ยนเป็น:**
```vue
<script setup lang="ts">
const { hero, socialLinks, siteSettings } = usePortfolioData()
</script>

<template>
  <!-- แทนที่ hardcode ด้วย: -->
  <span>{{ hero?.greeting }}</span>
  <h1>{{ hero?.full_name }}</h1>
  <p>{{ hero?.subtitle }}</p>

  <a :href="hero?.cta_primary_link">{{ hero?.cta_primary_text }}</a>
  <a :href="hero?.cta_secondary_link">{{ hero?.cta_secondary_text }}</a>

  <!-- Profile image -->
  <ProfileAvatar
    :src="hero ? `${apiBase}${siteSettings?.profile_image}` : '/images/profile.jpg'"
    :alt="hero?.full_name ?? 'Profile'"
  />
</template>
```

### 4.2 `SectionAbout.vue`

**เปลี่ยนจาก:**
- Title, bio paragraphs, personality tags, stats ที่ hardcode

**เปลี่ยนเป็น:**
```vue
<script setup lang="ts">
const { about } = usePortfolioData()
</script>

<template>
  <h2>{{ about?.title }}</h2>

  <p v-for="(paragraph, i) in about?.bio_paragraphs" :key="i">
    {{ paragraph }}
  </p>

  <span v-for="tag in about?.personality_tags" :key="tag">
    {{ tag }}
  </span>

  <div v-for="stat in about?.stats" :key="stat.label">
    <span>{{ stat.value }}</span>
    <span>{{ stat.label }}</span>
  </div>
</template>
```

### 4.3 `SectionSkills.vue`

**เปลี่ยนจาก:**
- `skills` array จาก `usePortfolioData()` (ไม่ต้องเปลี่ยน structure มาก)

**ตรวจสอบ:**
- ชื่อ field เปลี่ยนจาก camelCase เป็น snake_case (เช่น `sortOrder` → `sort_order`)
- property `category` ยังคงเหมือนเดิม

### 4.4 `SectionProjects.vue`

**เปลี่ยนจาก:**
- `projects` array จาก `usePortfolioData()`

**ตรวจสอบ field ที่เปลี่ยนชื่อ:**
| เดิม (camelCase) | ใหม่ (snake_case) |
|------------------|-------------------|
| `liveUrl` | `live_url` |
| `sourceUrl` | `source_url` |
| `sortOrder` | `sort_order` |

**รูปภาพ project:** ถ้ามี `image` ให้ใช้ URL จาก API:
```typescript
const imageUrl = project.image ? `${apiBase}${project.image}` : null
```

### 4.5 `SectionExperience.vue`

**เปลี่ยนจาก:**
- `experiences` array จาก `usePortfolioData()`

**ไม่ต้องเปลี่ยน field names มาก** — `role`, `company`, `period`, `description`, `highlights` ยังเหมือนเดิม
เพิ่ม `sort_order` field

### 4.6 `SectionContact.vue`

**เพิ่มการส่ง form ไปยัง API:**

```vue
<script setup lang="ts">
const config = useRuntimeConfig()
const apiBase = config.public.apiBaseUrl

const form = reactive({
  name: '',
  email: '',
  subject: '',
  message: '',
})

const submitting = ref(false)
const submitSuccess = ref(false)
const submitError = ref('')

const handleSubmit = async () => {
  submitting.value = true
  submitError.value = ''

  try {
    await $fetch(`${apiBase}/api/v1/contact`, {
      method: 'POST',
      body: form,
    })
    submitSuccess.value = true
    Object.assign(form, { name: '', email: '', subject: '', message: '' })
  } catch (err: any) {
    submitError.value = err?.data?.error || 'Failed to send message'
  } finally {
    submitting.value = false
  }
}
</script>
```

### 4.7 `TheNavbar.vue`

**เปลี่ยนจาก:**
- `navItems` hardcode

**เปลี่ยนเป็น:**
```vue
<script setup lang="ts">
const { navItems, siteSettings } = usePortfolioData()
</script>

<template>
  <!-- Brand text -->
  <span>{{ siteSettings?.site_title ?? 'Portfolio' }}</span>

  <!-- Navigation -->
  <a v-for="item in navItems" :key="item.href" :href="item.href">
    {{ item.label }}
  </a>
</template>
```

### 4.8 `TheFooter.vue`

**เปลี่ยนจาก:**
- Brand text และ tagline ที่ hardcode
- `socialLinks` จาก `usePortfolioData()`

**เปลี่ยนเป็น:**
```vue
<script setup lang="ts">
const { socialLinks, siteSettings } = usePortfolioData()
</script>

<template>
  <span>{{ siteSettings?.site_title ?? 'Portfolio' }}</span>
  <p>{{ siteSettings?.footer_tagline ?? 'Crafting digital experiences' }}</p>
</template>
```

---

## 5. แก้ไข `composables/useTheme.ts`

**เพิ่มการอ่าน default theme จาก API:**

```typescript
export const useTheme = () => {
  const { siteSettings } = usePortfolioData()

  const theme = useState<'midnight' | 'sunshine'>('theme', () => {
    if (import.meta.client) {
      const saved = localStorage.getItem('portfolio-theme')
      if (saved === 'midnight' || saved === 'sunshine') return saved
    }
    // ใช้ค่าจาก API เป็น default แทน hardcode
    return siteSettings.value?.default_theme ?? 'midnight'
  })

  // ... rest of existing logic
}
```

---

## 6. แก้ไข `pages/index.vue`

**ใช้ SEO metadata จาก API:**

```vue
<script setup lang="ts">
const { siteSettings } = usePortfolioData()

useHead({
  title: siteSettings.value?.page_title ?? 'Portfolio | Full-Stack Developer',
  meta: [
    {
      name: 'description',
      content: siteSettings.value?.meta_description ?? 'Full-Stack Developer portfolio',
    },
  ],
})
</script>
```

---

## 7. แก้ไข `types/portfolio.ts`

**เพิ่ม/แก้ไข interfaces ให้ตรงกับ API response:**

```typescript
// เพิ่ม field ที่มาจาก API
interface Skill {
  id: string
  name: string
  icon: string
  category: 'frontend' | 'backend' | 'devops' | 'tools'
  sort_order: number
  created_at: string
  updated_at: string
}

interface Project {
  id: string
  title: string
  description: string
  tags: string[]
  image?: string
  live_url?: string     // เปลี่ยนจาก liveUrl
  source_url?: string   // เปลี่ยนจาก sourceUrl
  sort_order: number
  created_at: string
  updated_at: string
}

interface Experience {
  id: string
  role: string
  company: string
  period: string
  description: string
  highlights: string[]
  sort_order: number
  created_at: string
  updated_at: string
}

interface SocialLink {
  id: string
  name: string
  url: string
  icon: string
  sort_order: number
  created_at: string
  updated_at: string
}
```

---

## 8. จัดการรูปภาพ

### Profile Image
- ปัจจุบัน: `/public/images/profile.jpg` (static file)
- ใหม่: ดึง path จาก `siteSettings.profile_image` แล้ว prefix ด้วย API base URL
- ตัวอย่าง: `http://localhost:8080/uploads/abc123.jpg`

### Project Images
- ปัจจุบัน: ยังไม่มี (ใช้ icon placeholder)
- ใหม่: ดึง path จาก `project.image` แล้ว prefix ด้วย API base URL
- ถ้า `image` เป็น empty string หรือ null ให้แสดง icon placeholder เหมือนเดิม

```typescript
const getImageUrl = (path?: string) => {
  if (!path) return null
  const config = useRuntimeConfig()
  return `${config.public.apiBaseUrl}${path}`
}
```

---

## 9. API Endpoints Reference

### Public Endpoints (ไม่ต้อง auth)

| Method | URL | คำอธิบาย |
|--------|-----|----------|
| `GET` | `/api/v1/portfolio` | ดึงข้อมูล portfolio ทั้งหมดใน response เดียว |
| `POST` | `/api/v1/contact` | ส่งข้อความจาก contact form |

### Public Response Format

```json
{
  "data": {
    "site_settings": { ... },
    "hero": { ... },
    "about": { ... },
    "skills": [ ... ],
    "projects": [ ... ],
    "experiences": [ ... ],
    "social_links": [ ... ],
    "nav_items": [ ... ]
  }
}
```

### Contact Form Request Body

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "subject": "Project Discussion",
  "message": "I'd like to discuss a project..."
}
```

---

## 10. Checklist ก่อน Deploy

- [ ] ตั้งค่า `NUXT_PUBLIC_API_BASE_URL` ใน `.env` ให้ชี้ไปที่ Admin API
- [ ] ลบข้อมูล hardcode ทั้งหมดออกจาก `usePortfolioData.ts`
- [ ] แก้ไข field names ใน components ให้ตรงกับ API response (snake_case)
- [ ] ทดสอบ contact form ว่าส่งข้อมูลไปยัง API ได้
- [ ] ทดสอบว่า profile image แสดงจาก API URL
- [ ] ทดสอบ default theme ว่าอ่านจาก API ถูกต้อง
- [ ] ทดสอบ SEO metadata ว่าอ่านจาก API ถูกต้อง
- [ ] ตั้งค่า CORS ที่ฝั่ง Admin API ให้รองรับ domain ของ Portfolio Frontend

---

> **หมายเหตุ:** API response ใช้ JSON envelope format: `{ "data": ..., "error": ..., "meta": ... }`
> ดังนั้นต้อง access ข้อมูลผ่าน `.data` property ของ response
