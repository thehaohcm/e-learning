<template>
  <div class="py-1">
    <h2 class="fw-bold mb-1"><i class="bi bi-journal-richtext"></i> Case Study — Quản lý dự án</h2>
    <p class="text-muted small mb-3">
      4 câu chuyện tình huống phủ trọn quy trình PMP từ đầu đến cuối: Predictive, Agile, Hybrid &amp; Con người,
      và Khủng hoảng &amp; Đạo đức. Đọc theo chương để nắm mạch quy trình, chú ý các <b>concept</b> và <b>bẫy thi</b>.
    </p>

    <!-- story selector -->
    <div class="d-flex flex-wrap gap-2 mb-3">
      <button v-for="s in stories" :key="s.id" type="button" class="btn btn-sm text-start"
        :class="s.id === active.id ? 'btn-primary' : 'btn-outline-primary'" @click="selectStory(s)">
        <b>Story {{ s.id }}</b> · {{ s.name }}
      </button>
    </div>

    <!-- active story header -->
    <div class="card border-0 shadow-sm mb-3">
      <div class="card-body">
        <div class="d-flex flex-wrap align-items-center gap-2 mb-2">
          <h4 class="fw-bold mb-0">{{ active.name }}</h4>
          <span class="badge rounded-pill bg-primary-subtle text-primary-emphasis">{{ active.world }}</span>
          <span class="badge rounded-pill bg-light text-dark border ms-auto">{{ active.chapters.length }} chương</span>
        </div>
        <p class="mb-0 cs-intro">{{ active.intro }}</p>
        <div class="mt-3">
          <div class="small text-muted mb-1"><i class="bi bi-headphones"></i> Nghe bản tóm tắt &amp; phân tích (audio NotebookLM)</div>
          <audio controls preload="none" class="w-100" :src="'/story-audio/story_' + active.id + '.m4a'"></audio>
        </div>
      </div>
    </div>

    <!-- template downloads for this story -->
    <div v-if="templatesFor(active.id).length" class="card border-0 shadow-sm mb-3">
      <div class="card-body">
        <h6 class="fw-bold mb-2"><i class="bi bi-download"></i> Template chứng từ dùng trong story này</h6>
        <p class="small text-muted mb-2">Tải về (.docx / .xlsx) để điền theo tình huống. EVM Tracker có sẵn công thức.</p>
        <div class="d-flex flex-wrap gap-2">
          <a v-for="t in templatesFor(active.id)" :key="t.file" :href="'/templates/' + t.file" :download="t.file"
            class="btn btn-sm btn-outline-secondary">
            <i :class="'bi me-1 ' + (t.file.endsWith('.xlsx') ? 'bi-file-earmark-spreadsheet text-success' : 'bi-file-earmark-word text-primary')"></i>{{ t.label }}
          </a>
        </div>
      </div>
    </div>

    <!-- chapter navigator -->
    <div class="d-flex flex-wrap gap-1 mb-3">
      <button v-for="c in active.chapters" :key="c.num" type="button"
        class="btn btn-sm py-0" :class="openCh === c.num ? 'btn-dark' : 'btn-outline-secondary'"
        @click="openCh = c.num">{{ c.num }}</button>
      <button type="button" class="btn btn-sm py-0 btn-outline-secondary ms-1" @click="expandAll = !expandAll">
        {{ expandAll ? 'Thu gọn' : 'Mở tất cả' }}
      </button>
    </div>

    <!-- chapters -->
    <div class="accordion cs-acc">
      <div v-for="c in active.chapters" :key="c.num" class="accordion-item border-0 shadow-sm mb-2 rounded overflow-hidden">
        <h2 class="accordion-header">
          <button class="accordion-button" :class="{ collapsed: !isOpen(c.num) }" type="button" @click="toggle(c.num)">
            <span class="badge bg-primary me-2 flex-shrink-0">Chương {{ c.num }}</span>
            <span class="fw-semibold">{{ c.title }}</span>
          </button>
        </h2>
        <div class="accordion-collapse collapse" :class="{ show: isOpen(c.num) }">
          <div class="accordion-body">
            <p class="cs-narr">{{ c.narrative }}</p>

            <div v-if="c.concepts.length">
              <div class="small text-uppercase text-muted fw-semibold mb-2">
                <i class="bi bi-lightbulb"></i> Concept PMP &amp; bẫy thi
              </div>
              <div v-for="(cc, i) in c.concepts" :key="i" class="cs-concept">
                <div>
                  <span v-if="cc.term" class="fw-bold text-primary">{{ cc.term }}</span>
                  <span v-if="cc.term && mainText(cc.desc)"> — </span>{{ mainText(cc.desc) }}
                </div>
                <div v-if="trapText(cc.desc)" class="cs-trap">
                  <i class="bi bi-exclamation-triangle-fill"></i> <b>Bẫy thi:</b> {{ trapText(cc.desc) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- extra / Phần 2 -->
    <div v-for="(ex, i) in active.extra" :key="'ex' + i" class="card border-0 shadow-sm mt-3">
      <div class="card-body">
        <h5 class="fw-bold mb-2"><i class="bi bi-collection"></i> {{ ex.heading }}</h5>
        <p v-for="(l, j) in ex.body" :key="j" class="mb-1 small cs-extra-line">{{ l }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { pmpStories, type PmpStory } from '../data/pmpStories'
import { setSubject, subjectSlug } from '../composables/useSubject'

// Narrative PMP case studies (read-only): pick a story, read chapter by chapter,
// each chapter shows the narrative plus the PMP concepts it triggers, with any
// "Bẫy thi" (exam trap) called out separately. Data: src/data/pmpStories.ts.
const stories = pmpStories
const active = ref<PmpStory>(stories[0])
const openCh = ref<number>(stories[0].chapters[0]?.num ?? 1)
const expandAll = ref(false)

// Case Study is PMP-only content; if opened under another topic prefix
// (e.g. /pentest/cases via the topic switcher), snap back to PMP.
const router = useRouter()
onMounted(() => {
  if (subjectSlug.value !== 'pmp') {
    setSubject('pmp')
    router.replace('/pmp/cases')
  }
})

const storyTemplates: Record<number, { label: string; file: string }[]> = {
  1: [
    { label: 'Project Charter', file: '13_Project_Charter.docx' },
    { label: 'Business Case', file: '14_Business_Case.docx' },
    { label: 'Scope Statement', file: '15_Scope_Statement.docx' },
    { label: 'WBS + Dictionary', file: '07_WBS_and_Dictionary.xlsx' },
    { label: 'Requirements Traceability', file: '08_Requirements_Traceability_Matrix.xlsx' },
    { label: 'Stakeholder Register', file: '01_Stakeholder_Register.xlsx' },
    { label: 'Risk Register', file: '02_Risk_Register.xlsx' },
    { label: 'Assumption Log', file: '05_Assumption_Log.xlsx' },
    { label: 'Communications Plan', file: '16_Communications_Plan.docx' },
    { label: 'RACI Matrix', file: '06_RACI_Matrix.xlsx' },
    { label: 'Procurement SOW', file: '17_Procurement_SOW.docx' },
    { label: 'EVM Tracker', file: '09_EVM_Tracker.xlsx' },
    { label: 'Change Log', file: '04_Change_Log.xlsx' },
    { label: 'Issue Log', file: '03_Issue_Log.xlsx' },
    { label: 'PM Plan', file: '18_Project_Management_Plan.docx' },
    { label: 'Closure Report', file: '19_Project_Closure_Report.docx' },
    { label: 'Lessons Learned', file: '12_Lessons_Learned_Register.xlsx' },
  ],
  2: [
    { label: 'Product Vision', file: '20_Product_Vision.docx' },
    { label: 'Product Backlog', file: '10_Product_Backlog.xlsx' },
    { label: 'Sprint Backlog', file: '11_Sprint_Backlog.xlsx' },
    { label: 'Definition of Done', file: '21_Definition_of_Done.docx' },
    { label: 'Lessons Learned', file: '12_Lessons_Learned_Register.xlsx' },
  ],
  3: [
    { label: 'PM Plan (Tailoring)', file: '18_Project_Management_Plan.docx' },
    { label: 'Communications Plan', file: '16_Communications_Plan.docx' },
    { label: 'RACI Matrix', file: '06_RACI_Matrix.xlsx' },
    { label: 'Stakeholder Register', file: '01_Stakeholder_Register.xlsx' },
    { label: 'Risk Register', file: '02_Risk_Register.xlsx' },
    { label: 'Product Backlog', file: '10_Product_Backlog.xlsx' },
    { label: 'Sprint Backlog', file: '11_Sprint_Backlog.xlsx' },
    { label: 'Lessons Learned', file: '12_Lessons_Learned_Register.xlsx' },
  ],
  4: [
    { label: 'EVM Tracker', file: '09_EVM_Tracker.xlsx' },
    { label: 'Change Log', file: '04_Change_Log.xlsx' },
    { label: 'Risk Register', file: '02_Risk_Register.xlsx' },
    { label: 'Issue Log', file: '03_Issue_Log.xlsx' },
    { label: 'Procurement SOW', file: '17_Procurement_SOW.docx' },
    { label: 'Closure Report', file: '19_Project_Closure_Report.docx' },
    { label: 'Stakeholder Register', file: '01_Stakeholder_Register.xlsx' },
    { label: 'Lessons Learned', file: '12_Lessons_Learned_Register.xlsx' },
  ],
}
function templatesFor(id: number) {
  return storyTemplates[id] ?? []
}

function selectStory(s: PmpStory) {
  active.value = s
  openCh.value = s.chapters[0]?.num ?? 1
  expandAll.value = false
}
function isOpen(n: number): boolean {
  return expandAll.value || openCh.value === n
}
function toggle(n: number) {
  if (expandAll.value) return
  openCh.value = openCh.value === n ? -1 : n
}

const TRAP = /B[aẫ]y thi\s*:/i
function mainText(desc: string): string {
  const idx = desc.search(TRAP)
  return (idx < 0 ? desc : desc.slice(0, idx)).trim()
}
function trapText(desc: string): string {
  const idx = desc.search(TRAP)
  if (idx < 0) return ''
  return desc.slice(idx).replace(TRAP, '').trim()
}
</script>

<style scoped>
.cs-intro { line-height: 1.6; }
.cs-narr { line-height: 1.7; white-space: pre-line; margin-bottom: 0.75rem; }
.cs-concept { padding: 0.5rem 0.75rem; margin-bottom: 0.5rem; background: var(--bs-light, #f8f9fa); border-left: 3px solid var(--bs-primary, #0d6efd); border-radius: 4px; line-height: 1.6; }
.cs-trap { margin-top: 0.4rem; padding: 0.4rem 0.6rem; background: #fff3cd; border-radius: 4px; color: #664d03; font-size: 0.92em; }
.cs-extra-line { line-height: 1.6; }
.accordion-button:not(.collapsed) { background-color: var(--bs-primary-bg-subtle, #cfe2ff); }
</style>
