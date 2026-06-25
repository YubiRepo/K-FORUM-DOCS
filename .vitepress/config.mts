import { defineConfig } from 'vitepress'
import fs from 'node:fs'
import path from 'node:path'

const ROOT = path.resolve(__dirname, '..')

// Folder yang TIDAK ikut dirender / di-scan
const IGNORE = new Set([
  '.git',
  '.github',
  '.vitepress',
  'node_modules',
  'Credentials',
  'public',
])

// FULL-AUTO: semua folder top-level kedeteksi sendiri.
// ORDER = urutan pilihan untuk folder yang sudah dikenal; folder lain
// (termasuk yang baru ditambah) otomatis ikut, diurutkan abjad setelahnya.
const ORDER = [
  'Modules',
  'API SPEC',
  'schemas',
  'flows',
  'Domain',
  'plans',
  'scrapper engine news',
  'deeplink',
  'bug testcase',
]

// Label cantik (emoji/nama) — opsional. Folder yang tidak ada di sini
// otomatis pakai nama foldernya apa adanya.
const LABELS: Record<string, string> = {
  Modules: '📦 Modules',
  'API SPEC': '🔌 API Spec',
  schemas: '🗄️ Database Schemas',
  flows: '🔁 Flows',
  Domain: '🧩 Domain',
  plans: '🗺️ Plans / Roadmap',
  'scrapper engine news': '📰 News Scraper',
  deeplink: '🔗 Deeplink',
  'bug testcase': '🐞 Bug & Test Case',
  Mobile: '📱 Mobile',
  Web: '💻 Web / Backoffice',
}

/** Ambil judul dari H1 pertama; fallback ke nama file yang dirapikan. */
function titleOf(file: string): string {
  try {
    const raw = fs.readFileSync(file, 'utf-8')
    const m = raw.match(/^\s*#\s+(.+)$/m)
    if (m) return m[1].replace(/[`*_]/g, '').trim()
  } catch {}
  return path
    .basename(file, '.md')
    .replace(/_/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

/** Link VitePress relatif terhadap root, segmen di-encode agar aman utk spasi. */
function linkOf(absFile: string): string {
  const rel = path.relative(ROOT, absFile).replace(/\\/g, '/').replace(/\.md$/, '')
  return '/' + rel.split('/').map(encodeURIComponent).join('/')
}

function prettyFolder(name: string): string {
  return LABELS[name] ?? name
}

/** Bangun item sidebar secara rekursif dari sebuah folder. */
function buildTree(dir: string): any[] {
  const entries = fs.readdirSync(dir, { withFileTypes: true })

  const folders = entries
    .filter((e) => e.isDirectory() && !IGNORE.has(e.name))
    .sort((a, b) => a.name.localeCompare(b.name))

  const files = entries
    .filter(
      (e) =>
        e.isFile() &&
        e.name.toLowerCase().endsWith('.md') &&
        e.name.toLowerCase() !== 'index.md' &&
        e.name.toLowerCase() !== 'readme.md',
    )
    .sort((a, b) => a.name.localeCompare(b.name))

  const items: any[] = []

  // file dulu, lalu subfolder (subfolder bisa di-collapse)
  for (const f of files) {
    const abs = path.join(dir, f.name)
    items.push({ text: titleOf(abs), link: linkOf(abs) })
  }

  for (const sub of folders) {
    const abs = path.join(dir, sub.name)
    const children = buildTree(abs)
    if (children.length === 0) continue
    items.push({
      text: prettyFolder(sub.name),
      collapsed: true,
      items: children,
    })
  }

  return items
}

function buildSidebar(): any[] {
  // Scan SEMUA folder top-level secara otomatis (kecuali yang di-IGNORE).
  const dirs = fs
    .readdirSync(ROOT, { withFileTypes: true })
    .filter((e) => e.isDirectory() && !IGNORE.has(e.name) && !e.name.startsWith('.'))
    .map((e) => e.name)

  // Urutkan: folder dikenal sesuai ORDER dulu, sisanya abjad.
  dirs.sort((a, b) => {
    const ia = ORDER.indexOf(a)
    const ib = ORDER.indexOf(b)
    if (ia !== -1 && ib !== -1) return ia - ib
    if (ia !== -1) return -1
    if (ib !== -1) return 1
    return a.localeCompare(b)
  })

  const sidebar: any[] = []
  for (const dir of dirs) {
    const items = buildTree(path.join(ROOT, dir))
    if (items.length === 0) continue
    sidebar.push({ text: prettyFolder(dir), collapsed: false, items })
  }
  return sidebar
}

export default defineConfig({
  title: 'K-Forum Docs',
  description: 'Dokumentasi teknis K-Forum — API Spec, Modules, Schemas & Flows',
  lang: 'id-ID',
  // Project page GitHub Pages: https://YubiRepo.github.io/K-FORUM-DOCS/
  base: '/K-FORUM-DOCS/',
  cleanUrls: true,
  lastUpdated: true,
  ignoreDeadLinks: true,

  // Dokumen banyak memuat placeholder seperti <tipe>, <id>, <token>.
  // Matikan raw-HTML agar karakter < > diperlakukan sebagai teks biasa
  // (mencegah error "Element is missing end tag" tanpa harus ubah tiap file).
  markdown: { html: false },

  // Jangan render JSON/credential & sumber non-doc lainnya
  srcExclude: ['**/Credentials/**', '**/node_modules/**', '**/README.md'],

  themeConfig: {
    outline: { level: [2, 3], label: 'Di halaman ini' },

    nav: [
      { text: 'Beranda', link: '/' },
      { text: 'Modules', link: '/Modules/Community/COMMUNITY_RULES' },
      { text: 'API Spec', link: '/API%20SPEC/Web/DIRECTORY_API_SPEC_BACKOFFICE_V2' },
      { text: 'Schemas', link: '/schemas/COMMUNITY_DB_SCHEMA' },
    ],

    sidebar: buildSidebar(),

    search: { provider: 'local' },

    docFooter: { prev: 'Sebelumnya', next: 'Selanjutnya' },
    darkModeSwitchLabel: 'Tampilan',
    sidebarMenuLabel: 'Menu',
    returnToTopLabel: 'Kembali ke atas',
    lastUpdatedText: 'Terakhir diperbarui',
  },
})
